package nginx

import (
	"reflect"
	"testing"
)

func TestParseCollectsAliasesPathsRootsAndPorts(t *testing.T) {
	config := `# configuration file /etc/nginx/sites-enabled/demo:
server {
    listen 80;
    listen 443 ssl;
    server_name demo.local www.demo.local *.demo.local;
    root /var/www/demo;
}
`

	domains := Parse(config)
	if len(domains) != 1 {
		t.Fatalf("got %d domains, want 1", len(domains))
	}
	domain := domains[0]
	if domain.Name != "demo.local" {
		t.Errorf("domain = %q, want demo.local", domain.Name)
	}
	if !reflect.DeepEqual(domain.Aliases, []string{"www.demo.local", "*.demo.local"}) {
		t.Errorf("aliases = %#v", domain.Aliases)
	}
	if domain.ConfigPath != "/etc/nginx/sites-enabled/demo" {
		t.Errorf("config path = %q", domain.ConfigPath)
	}
	if domain.DocumentRoot != "/var/www/demo" {
		t.Errorf("document root = %q", domain.DocumentRoot)
	}
	if !reflect.DeepEqual(domain.Ports, []int{80, 443}) {
		t.Errorf("ports = %#v", domain.Ports)
	}
}
