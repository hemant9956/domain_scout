package caddy

import (
	"bufio"
	"domain-scout/internal/inspectors"
	"strings"
)

func Parse(text string) []inspectors.Domain {
	seen := map[string]bool{}
	var result []inspectors.Domain
	sc := bufio.NewScanner(strings.NewReader(text))
	for sc.Scan() {
		line := strings.TrimSpace(strings.SplitN(sc.Text(), "#", 2)[0])
		if line == "" || strings.HasPrefix(line, "{") || strings.HasPrefix(line, "}") || strings.Contains(line, " ") {
			continue
		}
		line = strings.TrimSuffix(line, "{")
		for _, host := range strings.Split(line, ",") {
			host = strings.TrimSpace(host)
			if host == "" || strings.HasPrefix(host, ":") || strings.HasPrefix(host, "http://") || strings.HasPrefix(host, "https://") || seen[host] {
				continue
			}
			if strings.Contains(host, ".") {
				seen[host] = true
				result = append(result, inspectors.Domain{Name: host, Source: "caddy"})
			}
		}
	}
	return result
}
