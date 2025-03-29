package main

import (
	"cmp"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"slices"

	"github.com/mark-adams/gcp-ip-list/pkg/gcp"
	"github.com/mark-adams/gcp-ip-list/pkg/output"
)

var (
	version = "dev"

	scope        = flag.String("scope", "", "The scope (organization, folder, or project) to search (i.e. projects/abc-123 or organizations/123456)")
	scopePattern = regexp.MustCompile(`^organizations/\d+$|^folders/\d+$|^projects/\S+$`)

	format = flag.String("format", "table", "The output format (csv, json, table, list)")

	public  = flag.Bool("public", false, "Include public IPs only")
	private = flag.Bool("private", false, "Include private IPs only")

	showVersion = flag.Bool("version", false, "Display the current version")
)

func main() {
	log.SetFlags(0)
	flag.Parse()

	if *showVersion {
		fmt.Printf("gcp-ip-list %s\n", version)
		os.Exit(0)
	}

	if *scope == "" {
		log.Printf("error: scope flag is required (organizations/1234, folders/1234, or projects/1234)")
		os.Exit(1)
	}
	if !scopePattern.MatchString(*scope) {
		log.Fatalf("error: invalid scope: %s, scope must be organizations/1234, folders/1234, projects/1234", *scope)
	}

	if *public && *private {
		log.Fatalf("error: cannot specify both public and private flags")
	}

	formatters := output.GetFormatters()

	formatter := formatters[*format]
	if formatter == nil {
		log.Fatalf("error: invalid formatter: %s", *format)
	}

	ctx := context.Background()

	addresses, err := gcp.GetAllAddressesFromAssetInventory(ctx, *scope)
	if err != nil {
		log.Fatalf("error: failed to get addresses: %s", err)
	}

	if *public {
		addresses = gcp.FilterPublicAddresses(addresses)
	} else if *private {
		addresses = gcp.FilterPrivateAddresses(addresses)
	}

	// Sort the output by the address type (descending), then by resource type, then by resource name
	// (chosen somewhat arbitrarily)
	slices.SortFunc(addresses, func(a, b *gcp.Address) int {
		return cmp.Or(
			cmp.Compare(a.AddressType, b.AddressType)*-1,
			cmp.Compare(a.ResourceType, b.ResourceType),
			cmp.Compare(a.ResourceName, b.ResourceName),
		)
	})

	if err := formatter(os.Stdout, addresses); err != nil {
		log.Fatalf("error writing output: %s", err)
	}
}
