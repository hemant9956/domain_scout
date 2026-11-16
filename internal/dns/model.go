package dns

type Result struct {
	ResolvedIPs   []string `json:"resolved_ips"`
	MatchesServer bool     `json:"matches_server"`
	Status        string   `json:"status"`
}
