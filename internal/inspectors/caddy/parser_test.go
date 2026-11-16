package caddy

import (
	"reflect"
	"testing"
)

func TestParseExtractsSiteMetadata(t *testing.T) {
	domains := Parse(`demo.local, www.demo.local {
    root * /var/www/demo
    file_server
}`)
	if len(domains) != 1 {
		t.Fatalf("got %d domains, want 1", len(domains))
	}
	domain := domains[0]
	if domain.Name != "demo.local" {
		t.Errorf("domain = %q", domain.Name)
	}
	if !reflect.DeepEqual(domain.Aliases, []string{"www.demo.local"}) {
		t.Errorf("aliases = %#v", domain.Aliases)
	}
	if domain.DocumentRoot != "/var/www/demo" {
		t.Errorf("document root = %q", domain.DocumentRoot)
	}
	if !reflect.DeepEqual(domain.Ports, []int{80, 443}) {
		t.Errorf("ports = %#v", domain.Ports)
	}
}
