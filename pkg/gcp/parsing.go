package gcp

import (
	"strings"

	assetpb "cloud.google.com/go/asset/apiv1/assetpb"
)

type AddressGetter func(resource *assetpb.ResourceSearchResult) []*Address

const (
	AssetTypeComputeInstance       = "compute.googleapis.com/Instance"
	AssetTypeComputeAddress        = "compute.googleapis.com/Address"
	AssetTypeCloudSQLInstance      = "sqladmin.googleapis.com/Instance"
	AssetTypeContainerCluster      = "container.googleapis.com/Cluster"
	AssetTypeComputeForwardingRule = "compute.googleapis.com/ForwardingRule"
	AssetTypeComputeRouter         = "compute.googleapis.com/Router"
)

var getAddressByAssetType = map[string]AddressGetter{
	AssetTypeComputeInstance:       getAddressForGCEInstance,
	AssetTypeComputeAddress:        getAddressForAddress,
	AssetTypeCloudSQLInstance:      getAddressForSQLInstances,
	AssetTypeContainerCluster:      getAddressForGKECluster,
	AssetTypeComputeForwardingRule: getAddressForForwardingRule,
	AssetTypeComputeRouter:         getAddressForRouter,
}

func getAddressForGCEInstance(resource *assetpb.ResourceSearchResult) []*Address {
	ipStrings := []string{}

	externalIPs := resource.AdditionalAttributes.Fields["externalIPs"]
	if externalIPs != nil {
		ipList := externalIPs.GetListValue()
		if ipList == nil || len(ipList.Values) == 0 {
			return nil
		} else {
			for _, ip := range ipList.Values {
				ipVal := ip.GetStringValue()
				if ip.GetStringValue() != "" {
					ipStrings = append(ipStrings, ipVal)
				}
			}
		}
	}

	internalIPs := resource.AdditionalAttributes.Fields["internalIPs"]
	if internalIPs != nil {
		ipList := internalIPs.GetListValue()
		if ipList == nil || len(ipList.Values) == 0 {
			return nil
		} else {
			for _, ip := range ipList.Values {
				ipVal := ip.GetStringValue()
				if ip.GetStringValue() != "" {
					ipStrings = append(ipStrings, ipVal)
				}
			}
		}
	}

	addresses := []*Address{}
	for _, ip := range ipStrings {
		addresses = append(addresses, &Address{
			Address:      ip,
			ResourceName: resource.Name,
			AddressType:  ipType(ip),
			ResourceType: resource.AssetType,
		})
	}

	return addresses
}

func getAddressForAddress(resource *assetpb.ResourceSearchResult) []*Address {
	if resource.State != "IN_USE" {
		return nil
	}

	address := resource.AdditionalAttributes.Fields["address"]
	if address == nil {
		return nil
	}

	addressStr := address.GetStringValue()
	if addressStr == "" {
		return nil
	}

	return []*Address{
		{
			Address:      addressStr,
			ResourceName: resource.Name,
			AddressType:  ipType(addressStr),
			ResourceType: resource.AssetType,
		},
	}
}

func getAddressForSQLInstances(resource *assetpb.ResourceSearchResult) []*Address {
	dbResources := resource.GetVersionedResources()
	if len(dbResources) == 0 {
		return nil
	}

	dbResource := dbResources[0]
	if dbResource == nil {
		return nil
	}

	dbResourceValues := dbResource.GetResource()
	if dbResourceValues == nil {
		return nil
	}

	addresses := dbResourceValues.GetFields()["ipAddresses"]
	addressesList := addresses.GetListValue()
	addressesListValues := addressesList.GetValues()

	publicAddresses := []*Address{}

	for _, address := range addressesListValues {
		addressFields := address.GetStructValue().GetFields()
		typeValue := addressFields["type"].GetStringValue()

		if typeValue == "OUTGOING" {
			continue
		}

		addressValue := addressFields["ipAddress"].GetStringValue()
		publicAddresses = append(publicAddresses, &Address{
			Address:      addressValue,
			ResourceName: resource.Name,
			AddressType:  ipType(addressValue),
			ResourceType: resource.AssetType,
		})

	}

	return publicAddresses
}

func getAddressForGKECluster(resource *assetpb.ResourceSearchResult) []*Address {
	clusterResources := resource.GetVersionedResources()
	if len(clusterResources) == 0 {
		return nil
	}

	clusterResource := clusterResources[0]
	if clusterResource == nil {
		return nil
	}

	clusterResourceValues := clusterResource.GetResource()
	if clusterResourceValues == nil {
		return nil
	}

	privateClusterConfig := clusterResourceValues.GetFields()["privateClusterConfig"].GetStructValue()
	ipStrings := []string{}

	publicEndpoint := privateClusterConfig.GetFields()["publicEndpoint"]

	if publicEndpoint != nil && publicEndpoint.GetStringValue() != "" {
		ipStrings = append(ipStrings, publicEndpoint.GetStringValue())
	}

	privateEndpoint := privateClusterConfig.GetFields()["privateEndpoint"]

	if privateEndpoint != nil && privateEndpoint.GetStringValue() != "" {
		ipStrings = append(ipStrings, privateEndpoint.GetStringValue())
	}

	addresses := []*Address{}

	for _, ip := range ipStrings {
		addresses = append(addresses, &Address{
			Address:      ip,
			ResourceName: resource.Name,
			AddressType:  ipType(ip),
			ResourceType: resource.AssetType,
		})
	}

	return addresses
}

func getAddressForForwardingRule(resource *assetpb.ResourceSearchResult) []*Address {
	ruleResources := resource.GetVersionedResources()
	if len(ruleResources) == 0 {
		return nil
	}

	ruleResource := ruleResources[0]
	if ruleResource == nil {
		return nil
	}

	ruleResourceValues := ruleResource.GetResource()
	if ruleResourceValues == nil {
		return nil
	}

	addressField := ruleResourceValues.GetFields()["IPAddress"]

	if addressField == nil || addressField.GetStringValue() == "" {
		return nil
	}

	address := addressField.GetStringValue()

	return []*Address{
		{
			Address:      address,
			ResourceName: resource.Name,
			AddressType:  ipType(address),
			ResourceType: resource.AssetType,
		},
	}
}

func getAddressForRouter(resource *assetpb.ResourceSearchResult) []*Address {
	routerResources := resource.GetVersionedResources()
	if len(routerResources) == 0 {
		return nil
	}

	routerResource := routerResources[0]
	if routerResource == nil {
		return nil
	}

	routerResourceValues := routerResource.GetResource()
	if routerResourceValues == nil {
		return nil
	}

	natsField := routerResourceValues.GetFields()["nats"]

	if natsField == nil || natsField.GetListValue() == nil {
		return nil
	}

	natsList := natsField.GetListValue()
	natsListValues := natsList.GetValues()

	addresses := []*Address{}

	for _, nat := range natsListValues {
		natFields := nat.GetStructValue().GetFields()
		natIpsField := natFields["natIps"]
		if natIpsField == nil || natIpsField.GetListValue() == nil {
			continue
		}

		natIpsFieldList := natIpsField.GetListValue()
		for _, ip := range natIpsFieldList.GetValues() {
			ref := ip.GetStringValue()
			addresses = append(addresses, &Address{
				Address:      strings.Replace(ref, "https://www.googleapis.com/compute/v1/", "//compute.googleapis.com/", 1),
				ResourceName: resource.Name,
				AddressType:  AddressTypeReference,
				ResourceType: resource.AssetType,
			})
		}
	}

	return addresses
}
