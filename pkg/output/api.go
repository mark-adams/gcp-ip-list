package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"

	"github.com/mark-adams/gcp-ip-list/pkg/gcp"
	"github.com/olekukonko/tablewriter"
)

type FormatterFunc func(w io.Writer, addresses []*gcp.Address) error

func GetFormatters() map[string]FormatterFunc {
	return map[string]FormatterFunc{
		"csv":   OutputCSV,
		"json":  OutputJSON,
		"table": OutputTable,
		"list":  OutputList,
	}
}

// OutputJSON outputs the addresses as a JSON array
func OutputJSON(w io.Writer, addresses []*gcp.Address) error {
	enc := json.NewEncoder(w)
	addressWrapper := struct {
		Addresses []*gcp.Address `json:"addresses"`
	}{Addresses: addresses}

	if err := enc.Encode(addressWrapper); err != nil {
		return fmt.Errorf("error writing json: %w", err)
	}
	return nil
}

// OutputCSV outputs the addresses as a CSV with address, address_type, resource_type, and resource_name columns
func OutputCSV(w io.Writer, addresses []*gcp.Address) error {
	records := [][]string{
		{"address", "address_type", "resource_type", "resource_name"},
	}
	for _, address := range addresses {
		records = append(records, []string{address.Address, address.AddressType, address.ResourceType, address.ResourceName})
	}

	cw := csv.NewWriter(w)
	if err := cw.WriteAll(records); err != nil {
		return fmt.Errorf("error writing csv: %w", err)
	}

	cw.Flush()

	if err := cw.Error(); err != nil {
		return fmt.Errorf("error writing csv: %w", err)
	}

	return nil
}

// OutputTable outputs the IP addresses as a table with Address, Address Type, Resource Type, and Resource Name columns
func OutputTable(w io.Writer, addresses []*gcp.Address) error {
	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"Address", "Address Type", "Resource Type", "Resource Name"})

	for _, addr := range addresses {
		table.Append([]string{addr.Address, addr.AddressType, addr.ResourceType, addr.ResourceName})
	}

	table.Render()

	return nil
}

// OutputList outputs the IP addresses as a list, one per line
func OutputList(w io.Writer, addresses []*gcp.Address) error {
	for _, addr := range addresses {
		_, err := fmt.Fprintf(w, "%s\n", addr.Address)
		if err != nil {
			return err
		}
	}

	return nil
}
