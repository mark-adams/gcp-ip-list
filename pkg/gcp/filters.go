package gcp

// FilterPublicAddressees filters the given slice of addresses to only include public addresses
func FilterPublicAddresses(addrs []*Address) []*Address {
	filtered := []*Address{}

	for _, a := range addrs {
		if a.AddressType != AddressTypePublic {
			continue
		}
		filtered = append(filtered, a)
	}

	return filtered
}

// FilterPrivateAddressees filters the given slice of addresses to only include private addresses
func FilterPrivateAddresses(addrs []*Address) []*Address {
	filtered := []*Address{}

	for _, a := range addrs {
		if a.AddressType != AddressTypePrivate {
			continue
		}
		filtered = append(filtered, a)
	}

	return filtered
}
