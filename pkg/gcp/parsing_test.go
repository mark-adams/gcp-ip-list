package gcp_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"slices"
	"testing"

	assetpb "cloud.google.com/go/asset/apiv1/assetpb"
	"github.com/mark-adams/gcp-ip-list/pkg/gcp"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
)

const scope = "projects/fuzzy-pickles-428115"

type fakeAssetInventoryServer struct {
	assets []*assetpb.ResourceSearchResult

	assetpb.UnimplementedAssetServiceServer
}

func (f *fakeAssetInventoryServer) SearchAllResources(ctx context.Context, req *assetpb.SearchAllResourcesRequest) (*assetpb.SearchAllResourcesResponse, error) {
	response := assetpb.SearchAllResourcesResponse{}

	assetTypes := req.AssetTypes
	assets := []*assetpb.ResourceSearchResult{}

	for _, asset := range f.assets {
		if slices.Contains(assetTypes, asset.AssetType) {
			assets = append(assets, asset)
		}
	}

	response.Results = assets
	return &response, nil
}

func setupTestServer() (net.Listener, error) {
	f, err := os.Open("test-data/assets.json")
	if err != nil {
		return nil, fmt.Errorf("error opening asset test data file: %w", err)
	}
	defer f.Close() //nolint:errcheck

	assetBytes, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("error reading asset test data file: %w", err)
	}

	assetObjs := []json.RawMessage{}
	if err := json.Unmarshal(assetBytes, &assetObjs); err != nil {
		return nil, fmt.Errorf("error parsing asset test data: %w", err)
	}

	assets := []*assetpb.ResourceSearchResult{}

	for _, assetObj := range assetObjs {
		asset := assetpb.ResourceSearchResult{}
		if err := protojson.Unmarshal(assetObj, &asset); err != nil {
			return nil, fmt.Errorf("error parsing search result json: %w", err)
		}
		assets = append(assets, &asset)
	}

	assetInventoryServer := fakeAssetInventoryServer{assets: assets}
	gsrv := grpc.NewServer()
	assetpb.RegisterAssetServiceServer(gsrv, &assetInventoryServer)

	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return nil, fmt.Errorf("error setting up listener for test server: %w", err)
	}

	go func() {
		if err := gsrv.Serve(listener); err != nil {
			var nerr *net.OpError

			if errors.As(err, &nerr) {
				if nerr.Err.Error() == "use of closed network connection" {
					return
				}
			}

			panic(err)
		}
	}()

	return listener, nil
}

func TestGetAssets(t *testing.T) {
	testcases := []struct {
		name              string
		assetTypes        []string
		expectedAddresses []*gcp.Address
	}{
		{
			name:       "compute instances",
			assetTypes: []string{"compute.googleapis.com/Instance"},
			expectedAddresses: []*gcp.Address{
				{
					Address:      "34.83.128.26",
					AddressType:  "public",
					ResourceType: "compute.googleapis.com/Instance",
					ResourceName: "//compute.googleapis.com/projects/fuzzy-pickles-428115/zones/us-west1-a/instances/ip-list-test-vm",
				},
				{
					Address:      "10.0.3.2",
					AddressType:  "private",
					ResourceType: "compute.googleapis.com/Instance",
					ResourceName: "//compute.googleapis.com/projects/fuzzy-pickles-428115/zones/us-west1-a/instances/ip-list-test-vm",
				},
			},
		},
		{
			name:       "cloudsql_instances",
			assetTypes: []string{"sqladmin.googleapis.com/Instance"},
			expectedAddresses: []*gcp.Address{
				{
					Address:      "35.247.31.30",
					AddressType:  "public",
					ResourceName: "//cloudsql.googleapis.com/projects/fuzzy-pickles-428115/instances/ip-list-test-db",
					ResourceType: "sqladmin.googleapis.com/Instance",
				},
				{
					Address:      "10.252.0.3",
					AddressType:  "private",
					ResourceName: "//cloudsql.googleapis.com/projects/fuzzy-pickles-428115/instances/ip-list-test-db",
					ResourceType: "sqladmin.googleapis.com/Instance",
				},
			},
		},
		{
			name:       "load_balancers",
			assetTypes: []string{"compute.googleapis.com/ForwardingRule"},
			expectedAddresses: []*gcp.Address{
				{
					Address:      "34.54.244.120",
					AddressType:  "public",
					ResourceName: "//compute.googleapis.com/projects/fuzzy-pickles-428115/global/forwardingRules/ip-list-test-forwarding-rule-external-static",
					ResourceType: "compute.googleapis.com/ForwardingRule",
				},
				{
					Address:      "34.54.243.87",
					AddressType:  "public",
					ResourceName: "//compute.googleapis.com/projects/fuzzy-pickles-428115/global/forwardingRules/ip-list-test-forwarding-rule-external",
					ResourceType: "compute.googleapis.com/ForwardingRule",
				},
				{
					Address:      "10.0.2.2",
					AddressType:  "private",
					ResourceName: "//compute.googleapis.com/projects/fuzzy-pickles-428115/global/forwardingRules/ip-list-test-forwarding-rule-internal",
					ResourceType: "compute.googleapis.com/ForwardingRule",
				},
			},
		},
		{
			name:       "gke_cluster",
			assetTypes: []string{"container.googleapis.com/Cluster"},
			expectedAddresses: []*gcp.Address{
				{
					Address:      "34.105.114.31",
					AddressType:  "public",
					ResourceName: "//container.googleapis.com/projects/fuzzy-pickles-428115/locations/us-west1/clusters/ip-list-test-cluster",
					ResourceType: "container.googleapis.com/Cluster",
				},
				{
					Address:      "10.138.0.2",
					AddressType:  "private",
					ResourceName: "//container.googleapis.com/projects/fuzzy-pickles-428115/locations/us-west1/clusters/ip-list-test-cluster",
					ResourceType: "container.googleapis.com/Cluster",
				},
			},
		},
		{
			name:       "nat_router",
			assetTypes: []string{"compute.googleapis.com/Router"},
			expectedAddresses: []*gcp.Address{
				{
					Address:      "34.19.80.22",
					AddressType:  "public",
					ResourceName: "//compute.googleapis.com/projects/fuzzy-pickles-428115/regions/us-west1/routers/ip-list-test-router",
					ResourceType: "compute.googleapis.com/Router",
				},
			},
		},
		{
			name:       "addresses",
			assetTypes: []string{"compute.googleapis.com/Address"},
			expectedAddresses: []*gcp.Address{
				{
					Address:      "34.19.80.22",
					AddressType:  "public",
					ResourceName: "//compute.googleapis.com/projects/fuzzy-pickles-428115/regions/us-west1/addresses/ip-list-test-nat",
					ResourceType: "compute.googleapis.com/Address",
				},
				{
					Address:      "34.54.244.120",
					AddressType:  "public",
					ResourceName: "//compute.googleapis.com/projects/fuzzy-pickles-428115/global/addresses/ip-list-test-static-address",
					ResourceType: "compute.googleapis.com/Address",
				},
			},
		},
	}

	server, err := setupTestServer()
	if err != nil {
		t.Fatalf("error setting up test server: %s", err)
	}
	defer server.Close() //nolint:errcheck

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			addr, err := gcp.GetAddressesFromAssetInventory(
				context.Background(),
				scope,
				tc.assetTypes,

				// These are necessary to get the Google Cloud SDK to use the fake grpc server
				option.WithEndpoint(server.Addr().String()),
				option.WithoutAuthentication(),
				option.WithGRPCDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
			)
			if err != nil {
				t.Fatalf("error getting addresses from asset inventory: %s", err)
			}

			require.ElementsMatch(t, tc.expectedAddresses, addr)
		})
	}
}
