package nginx

import (
	"bufio"
	"regexp"
	"strings"

	"domain-scout/internal/inspectors"
)

var (
	serverBlock = regexp.MustCompile(`(?s)server\s*\{(.*?)\}`)
	serverName  = regexp.MustCompile(`(?m)^\s*server_name\s+([^;]+);`)
	root        = regexp.MustCompile(`(?m)^\s*root\s+([^;]+);`)
	proxy       = regexp.MustCompile(`(?m)^\s*proxy_pass\s+([^;]+);`)
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
		if m := serverName.FindStringSubmatch(block); len(m) > 1 {
			for _, name := range strings.Fields(m[1]) {
				if name == "_" || strings.HasPrefix(name, "~") {
					continue
				}
				d := inspectors.Domain{Name: name, Source: "nginx", ConfigPath: path, SSL: strings.Contains(block, "ssl")}
				if m := root.FindStringSubmatch(block); len(m) > 1 {
					d.DocumentRoot = strings.TrimSpace(m[1])
				}
				if m := proxy.FindStringSubmatch(block); len(m) > 1 {
					d.ProxyTarget = strings.TrimSpace(m[1])
				}
				domains = append(domains, d)
			}
		}
	}
	return domains
}
