package apache

import (
	"reflect"
	"testing"
)

func TestParseAndApplyDocumentRoots(t *testing.T) {
	output := `*:80 is a NameVirtualHost
         port 80 namevhost apache.com (/etc/apache2/sites-enabled/site.conf:1)
                 alias www.apache.com
         port 443 namevhost apache2.com (/etc/apache2/sites-enabled/site2.conf:1)
                 alias www.apache2.com
                 alias www2.apache2.com
`
	domains := Parse(output)
	configs := `# DOMAIN_SCOUT_CONFIG:/etc/apache2/sites-enabled/site.conf
<VirtualHost *:80>
  ServerName apache.com
  ServerAlias www.apache.com
  DocumentRoot /var/www/apache
</VirtualHost>
# DOMAIN_SCOUT_CONFIG:/etc/apache2/sites-enabled/site2.conf
<VirtualHost *:443>
  ServerName apache2.com
  ServerAlias www.apache2.com www2.apache2.com
  DocumentRoot /var/www/apache2
</VirtualHost>
`
	applyDocumentRoots(domains, configs)

	if len(domains) != 2 {
		t.Fatalf("got %d domains, want 2", len(domains))
	}
	if domains[0].Name != "apache.com" || domains[0].ConfigPath != "/etc/apache2/sites-enabled/site.conf" {
		t.Errorf("unexpected first domain: %#v", domains[0])
	}
	if !reflect.DeepEqual(domains[0].Aliases, []string{"www.apache.com"}) {
		t.Errorf("first aliases = %#v", domains[0].Aliases)
	}
	if !reflect.DeepEqual(domains[0].Ports, []int{80}) {
		t.Errorf("first ports = %#v", domains[0].Ports)
	}
	if domains[0].DocumentRoot != "/var/www/apache" {
		t.Errorf("first document root = %q", domains[0].DocumentRoot)
	}
	if !reflect.DeepEqual(domains[1].Aliases, []string{"www.apache2.com", "www2.apache2.com"}) {
		t.Errorf("second aliases = %#v", domains[1].Aliases)
	}
	if !reflect.DeepEqual(domains[1].Ports, []int{443}) {
		t.Errorf("second ports = %#v", domains[1].Ports)
	}
	if domains[1].DocumentRoot != "/var/www/apache2" {
		t.Errorf("second document root = %q", domains[1].DocumentRoot)
	}
}
