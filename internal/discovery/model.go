package discovery

import (
	"domain-scout/internal/dns"
	"domain-scout/internal/inspectors"
	"domain-scout/internal/server"
	"time"
)

type Result struct {
	Status           string         `json:"status"`
	ScannedAt        time.Time      `json:"scanned_at"`
	Server           server.Info    `json:"server"`
	DetectedServices []string       `json:"detected_services"`
	TotalDomains     int            `json:"total_domains"`
	Domains          []DomainResult `json:"domains"`
	Warnings         []string       `json:"warnings,omitempty"`
}
type DomainResult struct {
	inspectors.Domain
	DNS dns.Result `json:"dns"`
}
