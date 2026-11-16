package nginx

import (
	"bufio"
	"regexp"
	"strconv"
	"strings"

	"domain-scout/internal/inspectors"
)

var (
	serverBlock = regexp.MustCompile(`(?s)server\s*\{(.*?)\}`)
	serverName  = regexp.MustCompile(`(?m)^\s*server_name\s+([^;]+);`)
	root        = regexp.MustCompile(`(?m)^\s*root\s+([^;]+);`)
	proxy       = regexp.MustCompile(`(?m)^\s*proxy_pass\s+([^;]+);`)
	listen      = regexp.MustCompile(`(?m)^\s*listen\s+([^;]+);`)
)

func Parse(config string) []inspectors.Domain {
	var domains []inspectors.Domain
	var lines []string
	currentFile := ""
	fileByLine := map[int]string{}
	scanner := bufio.NewScanner(strings.NewReader(config))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "# configuration file ") && strings.HasSuffix(line, ":") {
			currentFile = strings.TrimSuffix(strings.TrimPrefix(line, "# configuration file "), ":")
		}
		lines = append(lines, line)
		fileByLine[len(lines)-1] = currentFile
	}
	text := strings.Join(lines, "\n")
	for _, loc := range serverBlock.FindAllStringSubmatchIndex(text, -1) {
		block := text[loc[2]:loc[3]]
		startLine := strings.Count(text[:loc[0]], "\n")
		path := fileByLine[startLine]
		names := serverName.FindStringSubmatch(block)
		if len(names) < 2 {
			continue
		}
		var hosts []string
		for _, name := range strings.Fields(names[1]) {
			if name != "_" && !strings.HasPrefix(name, "~") {
				hosts = append(hosts, name)
			}
		}
		if len(hosts) == 0 {
			continue
		}
		d := inspectors.Domain{
			Name: hosts[0], Source: "nginx", ConfigPath: path,
			Aliases: append([]string{}, hosts[1:]...),
			Ports:   parsePorts(block),
		}
		if len(d.Ports) == 0 {
			d.Ports = []int{80}
		}
		for _, port := range d.Ports {
			if port == 443 {
				d.SSL = true
			}
		}
		for _, line := range listen.FindAllStringSubmatch(block, -1) {
			if len(line) > 1 && strings.Contains(line[1], "ssl") {
				d.SSL = true
			}
		}
		if m := root.FindStringSubmatch(block); len(m) > 1 {
			d.DocumentRoot = strings.TrimSpace(m[1])
		}
		if m := proxy.FindStringSubmatch(block); len(m) > 1 {
			d.ProxyTarget = strings.TrimSpace(m[1])
		}
		domains = append(domains, d)
	}
	return domains
}

func parsePorts(block string) []int {
	var ports []int
	seen := map[int]bool{}
	for _, match := range listen.FindAllStringSubmatch(block, -1) {
		if len(match) < 2 {
			continue
		}
		value := strings.Fields(match[1])
		if len(value) == 0 {
			continue
		}
		address := strings.Trim(value[0], "[]")
		if i := strings.LastIndex(address, ":"); i >= 0 {
			address = address[i+1:]
		}
		port, err := strconv.Atoi(address)
		if err != nil || port < 1 || port > 65535 || seen[port] {
			continue
		}
		seen[port] = true
		ports = append(ports, port)
	}
	return ports
}
