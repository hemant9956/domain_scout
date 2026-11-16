package apache

import (
	"domain-scout/internal/inspectors"
	"regexp"
	"strconv"
	"strings"
)

var (
	vhost = regexp.MustCompile(`(?i)\bnamevhost\s+([^\s]+)\s+\((.+):\d+\)`)
	alias = regexp.MustCompile(`(?i)^\s*alias\s+([^\s]+)`)
	port  = regexp.MustCompile(`(?i)(?:\*:(\d+)|\bport\s+(\d+))`)
)

func Parse(text string) []inspectors.Domain {
	var result []inspectors.Domain
	var current *inspectors.Domain
	for _, line := range strings.Split(text, "\n") {
		if match := vhost.FindStringSubmatch(line); len(match) > 2 {
			domain := inspectors.Domain{
				Name: match[1], Source: "apache", ConfigPath: strings.TrimSpace(match[2]),
				Aliases: []string{}, Ports: []int{},
			}
			if matches := port.FindStringSubmatch(line); len(matches) > 2 {
				portNumber := matches[1]
				if portNumber == "" {
					portNumber = matches[2]
				}
				if value, err := strconv.Atoi(portNumber); err == nil && value > 0 && value <= 65535 {
					domain.Ports = append(domain.Ports, value)
				}
			}
			result = append(result, domain)
			current = &result[len(result)-1]
			continue
		}
		if current == nil {
			continue
		}
		if match := alias.FindStringSubmatch(line); len(match) > 1 {
			current.Aliases = append(current.Aliases, strings.TrimSpace(match[1]))
		}
	}
	return result
}
