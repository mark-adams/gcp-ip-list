package gcp

import (
	"context"
	"fmt"
	"net"
	"slices"

	asset "cloud.google.com/go/asset/apiv1"
	assetpb "cloud.google.com/go/asset/apiv1/assetpb"
	"golang.org/x/exp/maps"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

const (
	// AddressTypePublic refers to a public, internet-routable IP address
	AddressTypePublic = "public"

	// AddressTypePrivate refers to a private, non-internet-routable IP address
	AddressTypePrivate = "private"

	// Special placeholder to handle references to Address assets from NAT routers
	// This is needed because some assets only reference the Address resource name
	// that they are attached to instead of the IP address itself.
	// We use this placeholder to flag these assets so we can normalize them to the actual IP address
	// later. The output of the application should never show these.
	AddressTypeReference = "reference"
)

type Address struct {
	Address      string `json:"address"`
	AddressType  string `json:"type"`
	ResourceName string `json:"resource_name"`
	ResourceType string `json:"asset_type"`
}

// GetAllAddressesFromAssetInventory queries the Cloud Asset Inventory API and returns back IP addresses from all supported asset types
func GetAllAddressesFromAssetInventory(ctx context.Context, scope string, opts ...option.ClientOption) ([]*Address, error) {
	return GetAddressesFromAssetInventory(ctx, scope, maps.Keys(getAddressByAssetType), opts...)
}

// GetAddressesFromAssetInventory queries the Cloud Asset Inventory API and returns back IP addresses from the specified asset types
func GetAddressesFromAssetInventory(ctx context.Context, scope string, assetTypes []string, opts ...option.ClientOption) ([]*Address, error) {
	c, err := asset.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("error setting up client: %w", err)
	}
	defer c.Close() //nolint:errcheck

	for _, val := range assetTypes {
		if _, ok := getAddressByAssetType[val]; !ok {
			return nil, fmt.Errorf("unsupported asset type: %s", val)
		}
	}

	removeAddressesLater := false

	if slices.Contains(assetTypes, AssetTypeComputeRouter) && !slices.Contains(assetTypes, AssetTypeComputeAddress) {
		// NAT routers only reference the Address resource name, not the IP address itself.
		// As a result, if we are pulling down the NAT routers without pulling down the Address resources explicitly, we need to add it under the hood
		// so we can resolve the reference later
		assetTypes = append(assetTypes, AssetTypeComputeAddress)
		removeAddressesLater = true
	}

	req := &assetpb.SearchAllResourcesRequest{
		Scope:      scope,
		AssetTypes: assetTypes,
		ReadMask: &fieldmaskpb.FieldMask{
			Paths: []string{"*"},
		},
		PageSize: 500,
	}

	it := c.SearchAllResources(ctx, req)

	var results []*Address

	for {
		resource, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error searching for resources: %w", err)
		}

		addressGetter := getAddressByAssetType[resource.AssetType]
		if addressGetter == nil {
			return nil, fmt.Errorf("unexpected asset type: %s", resource.AssetType)
		}

		addresses := addressGetter(resource)
		results = append(results, addresses...)
	}

	return cleanupAssets(results, removeAddressesLater), nil
}

// cleanupAssets removes duplicate addresses from the list of assets
func cleanupAssets(assets []*Address, removeAddresses bool) []*Address {
	// Resolve references to the IP of their associated Address resource
	computeAddressMap := map[string]*Address{}
	routerReferences := []*Address{}

	for _, addr := range assets {
		if addr.ResourceType == AssetTypeComputeAddress {
			computeAddressMap[addr.ResourceName] = addr
		}

		if addr.AddressType == "reference" && addr.ResourceType == AssetTypeComputeRouter {
			routerReferences = append(routerReferences, addr)
		}
	}

	for _, router := range routerReferences {
		if addr, ok := computeAddressMap[router.Address]; ok {
			router.Address = addr.Address
			router.AddressType = addr.AddressType
		}
	}

	// Remove reference assets from the list of assets since they should have been resolved above
	assets = slices.DeleteFunc(assets, func(a *Address) bool {
		return a.AddressType == "reference"
	})

	// Iterate over the list of assets and remove duplicate Address entries where the IP matches another
	// more-specific asset.
	seen := map[string]*Address{}

	for _, asset := range assets {
		if match, ok := seen[asset.Address]; !ok {
			seen[asset.Address] = asset
		} else {
			if match.ResourceType == AssetTypeComputeAddress {
				seen[asset.Address] = asset
			}
		}
	}

	addresses := []*Address{}
	for _, asset := range seen {
		if removeAddresses && asset.ResourceType == AssetTypeComputeAddress {
			continue
		}
		addresses = append(addresses, asset)
	}

	return addresses
}

func ipType(ip string) string {
	ipAddr := net.ParseIP(ip)
	if ipAddr.IsPrivate() {
		return AddressTypePrivate
	} else {
		return AddressTypePublic
	}
}
