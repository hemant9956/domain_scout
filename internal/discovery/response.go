package discovery

type ScanResponse struct {
	Servers []ServerInfo `json:"servers"`
	System  SystemInfo   `json:"system"`
}

type ServerInfo struct {
	Server  string       `json:"server"`
	Running bool         `json:"running"`
	Domains []DomainInfo `json:"domains"`
}

type DomainInfo struct {
	Domain       string   `json:"domain"`
	Aliases      []string `json:"aliases"`
	ConfigFile   string   `json:"config_file"`
	DocumentRoot string   `json:"document_root"`
	Ports        []int    `json:"ports"`
}
