package gcp_test

import (
	"testing"

	"github.com/mark-adams/gcp-ip-list/pkg/gcp"
	"github.com/stretchr/testify/require"
)

func TestFilterPublicAddresses(t *testing.T) {
	testAddresses := []*gcp.Address{
		{
			Address:     "10.0.1.2",
			AddressType: gcp.AddressTypePrivate,
		},
		{
			Address:     "92.12.13.14",
			AddressType: gcp.AddressTypePublic,
		},
	}

	filtered := gcp.FilterPublicAddresses(testAddresses)

	require.Len(t, filtered, 1)
	require.EqualValues(t, testAddresses[1:2], filtered)
}

func TestFilterPrivateAddresses(t *testing.T) {
	testAddresses := []*gcp.Address{
		{
			Address:     "10.0.1.2",
			AddressType: gcp.AddressTypePrivate,
		},
		{
			Address:     "92.12.13.14",
			AddressType: gcp.AddressTypePublic,
		},
	}

	filtered := gcp.FilterPrivateAddresses(testAddresses)

	require.Len(t, filtered, 1)
	require.EqualValues(t, testAddresses[0:1], filtered)
}
