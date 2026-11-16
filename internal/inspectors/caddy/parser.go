package caddy

import (
	"bufio"
	"net"
	"strconv"
	"strings"

	"domain-scout/internal/inspectors"
)

func Parse(text string) []inspectors.Domain {
	var domains []inspectors.Domain
	var pending []string
	var pendingPorts []int
	var current []int
	var root string
	depth := 0
	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		line := strings.TrimSpace(strings.SplitN(scanner.Text(), "#", 2)[0])
		if line == "" {
			continue
		}
		if depth == 0 {
			if line == "{" {
				current = addDomains(&domains, pending, pendingPorts)
				pending = nil
				pendingPorts = nil
				depth = 1
				continue
			}
			header := line
			inlineBlock := strings.Contains(line, "{")
			if inlineBlock {
				header = strings.TrimSpace(strings.SplitN(line, "{", 2)[0])
			}
			if hosts, ports := parseSiteAddresses(header); len(hosts) > 0 {
				pending, pendingPorts = hosts, ports
			}
			if inlineBlock {
				current = addDomains(&domains, pending, pendingPorts)
				pending = nil
				pendingPorts = nil
				depth = 1
			}
			continue
		}

		if strings.HasPrefix(line, "root ") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				root = fields[len(fields)-1]
			}
		}
		depth += strings.Count(line, "{") - strings.Count(line, "}")
		if depth <= 0 {
			for _, index := range current {
				domains[index].DocumentRoot = root
			}
			current = nil
			root = ""
			depth = 0
		}
	}
	return domains
}

func parseSiteAddresses(header string) ([]string, []int) {
	header = strings.TrimSpace(header)
	if header == "" || strings.HasPrefix(header, "{") {
		return nil, nil
	}
	var hosts []string
	var ports []int
	for _, address := range strings.Split(header, ",") {
		address = strings.TrimSpace(address)
		if strings.ContainsAny(address, " \t") {
			return nil, nil
		}
		port := 0
		if strings.HasPrefix(address, "http://") {
			port = 80
		} else if strings.HasPrefix(address, "https://") {
			port = 443
		}
		address = strings.TrimPrefix(address, "http://")
		address = strings.TrimPrefix(address, "https://")
		if address == "" || strings.HasPrefix(address, ":") {
			continue
		}
		host := address
		if parsedHost, hostPort, err := net.SplitHostPort(address); err == nil {
			host = parsedHost
			port, _ = strconv.Atoi(hostPort)
		}
		if strings.Contains(host, ".") {
			hosts = append(hosts, host)
			if port == 0 {
				ports = appendPort(ports, 80)
				port = 80
				ports = appendPort(ports, 443)
			} else {
				ports = appendPort(ports, port)
			}
		}
	}
	return hosts, ports
}

func addDomains(domains *[]inspectors.Domain, hosts []string, ports []int) []int {
	if len(hosts) == 0 {
		return nil
	}
	indices := make([]int, 0, len(hosts))
	primary := inspectors.Domain{
		Name: hosts[0], Source: "caddy", ConfigPath: "/etc/caddy/Caddyfile",
		Aliases: append([]string{}, hosts[1:]...), Ports: append([]int{}, ports...),
	}
	*domains = append(*domains, primary)
	indices = append(indices, len(*domains)-1)
	return indices
}

func appendPort(ports []int, value int) []int {
	for _, port := range ports {
		if port == value {
			return ports
		}
	}
	return append(ports, value)
}
