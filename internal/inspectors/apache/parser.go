package apache

import (
	"domain-scout/internal/inspectors"
	"regexp"
	"strings"
)

var vhost = regexp.MustCompile(`(?i)namevhost\s+([^\s]+)|alias\s+([^\s]+)`)

func Parse(text string) []inspectors.Domain {
	seen := map[string]bool{}
	var result []inspectors.Domain
	for _, m := range vhost.FindAllStringSubmatch(text, -1) {
		name := m[1]
		if name == "" {
			name = m[2]
		}
		name = strings.TrimSpace(strings.TrimSuffix(name, ":"))
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		result = append(result, inspectors.Domain{Name: name, Source: "apache"})
	}
	return result
}
