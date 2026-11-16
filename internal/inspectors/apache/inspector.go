package apache

import (
	"context"
	"domain-scout/internal/inspectors"
	"domain-scout/internal/ssh"
	"fmt"
	"regexp"
	"strings"
)

type Inspector struct{}

func New() *Inspector             { return &Inspector{} }
func (i *Inspector) Name() string { return "apache" }
func (i *Inspector) Detect(ctx context.Context, ex ssh.Executor) (bool, error) {
	_, e1 := ex.Run(ctx, "command -v apachectl")
	if ctx.Err() != nil {
		return false, ctx.Err()
	}
	if e1 == nil {
		return true, nil
	}
	_, e2 := ex.Run(ctx, "command -v httpd")
	if ctx.Err() != nil {
		return false, ctx.Err()
	}
	return e2 == nil, nil
}
func (i *Inspector) Running(ctx context.Context, ex ssh.Executor) (bool, error) {
	for _, service := range []string{"apache2", "httpd"} {
		if _, err := ex.Run(ctx, "systemctl is-active --quiet "+service); err == nil {
			return true, nil
		}
		if ctx.Err() != nil {
			return false, ctx.Err()
		}
	}
	if ctx.Err() != nil {
		return false, ctx.Err()
	}
	if _, err := ex.Run(ctx, "ps -eo comm= | grep -Eq '^(apache2|httpd)$'"); err == nil {
		return true, nil
	}
	if ctx.Err() != nil {
		return false, ctx.Err()
	}
	return false, nil
}
func (i *Inspector) Discover(ctx context.Context, ex ssh.Executor) ([]inspectors.Domain, error) {
	out, err := ex.Run(ctx, "sudo -n apachectl -S 2>&1")
	if err != nil {
		out, err = ex.Run(ctx, "sudo -n httpd -S 2>&1")
	}
	if err != nil {
		return nil, err
	}
	domains := Parse(out)
	paths := make([]string, 0, len(domains))
	seen := map[string]bool{}
	for _, domain := range domains {
		if domain.ConfigPath != "" && !seen[domain.ConfigPath] {
			seen[domain.ConfigPath] = true
			paths = append(paths, domain.ConfigPath)
		}
	}
	if len(paths) == 0 {
		return domains, nil
	}
	script := "for f in " + shellPaths(paths) + "; do printf '\\n# DOMAIN_SCOUT_CONFIG:%s\\n' \"$f\"; cat -- \"$f\"; done"
	configs, configErr := ex.Run(ctx, "sudo -n sh -c "+shellQuote(script))
	if configErr != nil {
		configs, configErr = ex.Run(ctx, "sh -c "+shellQuote(script))
	}
	if configs != "" {
		applyDocumentRoots(domains, configs)
	}
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if configErr != nil {
		return domains, fmt.Errorf("read Apache virtual host configuration: %w", configErr)
	}
	return domains, nil
}

var configMarker = regexp.MustCompile(`^# DOMAIN_SCOUT_CONFIG:(.*)$`)

func applyDocumentRoots(domains []inspectors.Domain, text string) {
	roots := map[string]string{}
	currentFile := ""
	inVHost := false
	var names []string
	root := ""
	flush := func() {
		for _, name := range names {
			roots[currentFile+"\x00"+name] = root
		}
		names = nil
		root = ""
	}
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if marker := configMarker.FindStringSubmatch(line); len(marker) > 1 {
			if inVHost {
				flush()
			}
			currentFile = strings.TrimSpace(marker[1])
			inVHost = false
			continue
		}
		line = strings.TrimSpace(strings.SplitN(line, "#", 2)[0])
		lower := strings.ToLower(line)
		if strings.HasPrefix(lower, "<virtualhost ") {
			if inVHost {
				flush()
			}
			inVHost = true
			continue
		}
		if !inVHost {
			continue
		}
		if strings.HasPrefix(lower, "</virtualhost") {
			flush()
			inVHost = false
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		switch strings.ToLower(fields[0]) {
		case "servername":
			names = append(names, fields[1])
		case "serveralias":
			names = append(names, fields[1:]...)
		case "documentroot":
			root = strings.Trim(fields[1], `"'`)
		}
	}
	if inVHost {
		flush()
	}
	for i := range domains {
		if value := roots[domains[i].ConfigPath+"\x00"+domains[i].Name]; value != "" {
			domains[i].DocumentRoot = value
		}
	}
}

func shellPaths(paths []string) string {
	quoted := make([]string, 0, len(paths))
	for _, path := range paths {
		quoted = append(quoted, shellQuote(path))
	}
	return strings.Join(quoted, " ")
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}
