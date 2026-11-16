package discovery

import (
	"domain-scout/internal/inspectors"
	"strings"
)

func Normalize(input []inspectors.Domain) []inspectors.Domain {
	index := map[string]int{}
	out := make([]inspectors.Domain, 0, len(input))
	for _, d := range input {
		d.Name = strings.TrimSuffix(strings.ToLower(strings.TrimSpace(d.Name)), ".")
		if d.Name == "" || !strings.Contains(d.Name, ".") {
			continue
		}
		if pos, ok := index[d.Name]; ok {
			existing := &out[pos]
			if existing.Source != d.Source && !strings.Contains(existing.Source, d.Source) {
				existing.Source += "," + d.Source
			}
			if existing.ConfigPath == "" {
				existing.ConfigPath = d.ConfigPath
			}
			if existing.DocumentRoot == "" {
				existing.DocumentRoot = d.DocumentRoot
			}
			if existing.ProxyTarget == "" {
				existing.ProxyTarget = d.ProxyTarget
			}
			existing.SSL = existing.SSL || d.SSL
			existing.Aliases = appendUnique(existing.Aliases, d.Aliases...)
			existing.Ports = appendUniqueInts(existing.Ports, d.Ports...)
			continue
		}
		index[d.Name] = len(out)
		out = append(out, d)
	}
	return out
}

func appendUnique(values []string, added ...string) []string {
	seen := make(map[string]bool, len(values)+len(added))
	for _, value := range values {
		seen[value] = true
	}
	for _, value := range added {
		if value != "" && !seen[value] {
			values = append(values, value)
			seen[value] = true
		}
	}
	return values
}

func appendUniqueInts(values []int, added ...int) []int {
	seen := make(map[int]bool, len(values)+len(added))
	for _, value := range values {
		seen[value] = true
	}
	for _, value := range added {
		if value > 0 && !seen[value] {
			values = append(values, value)
			seen[value] = true
		}
	}
	return values
}
