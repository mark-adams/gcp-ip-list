package output_test

import (
	"bytes"
	"testing"

	"github.com/mark-adams/gcp-ip-list/pkg/gcp"
	"github.com/mark-adams/gcp-ip-list/pkg/output"
	"github.com/stretchr/testify/require"
)

var testAddresses = []*gcp.Address{
	{
		Address:      "1.2.3.4",
		AddressType:  gcp.AddressTypePublic,
		ResourceType: "compute.googleapis.com/Instance",
		ResourceName: "//compute.googleapis.com/instance-1",
	},
	{
		Address:      "5.6.7.8",
		AddressType:  gcp.AddressTypePublic,
		ResourceType: "sqladmin.googleapis.com/Instance",
		ResourceName: "//sqladmin.googleapis.com/instance-2",
	},
}

func TestOutputCSV(t *testing.T) {
	buf := bytes.NewBuffer(nil)

	err := output.OutputCSV(buf, testAddresses)
	require.NoError(t, err)

	output := buf.String()
	require.Equal(t, "address,address_type,resource_type,resource_name\n1.2.3.4,public,compute.googleapis.com/Instance,//compute.googleapis.com/instance-1\n5.6.7.8,public,sqladmin.googleapis.com/Instance,//sqladmin.googleapis.com/instance-2\n", output)
}

func TestOutputList(t *testing.T) {
	buf := bytes.NewBuffer(nil)

	err := output.OutputList(buf, testAddresses)
	require.NoError(t, err)

	output := buf.String()
	require.Equal(t, "1.2.3.4\n5.6.7.8\n", output)
}
