package discovery

import "domain-scout/internal/inspectors"

func Aggregate(groups ...[]inspectors.Domain) []inspectors.Domain {
	var all []inspectors.Domain
	for _, g := range groups {
		all = append(all, g...)
	}
	return Normalize(all)
}
