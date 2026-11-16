package dns

import (
	"context"
	"net"
)

type Verifier struct{ resolver *Resolver }

func NewVerifier() *Verifier { return &Verifier{resolver: NewResolver()} }
func (v *Verifier) Verify(ctx context.Context, domain, serverIP string) Result {
	ips, err := v.resolver.LookupIP(ctx, domain)
	if err != nil {
		return Result{ResolvedIPs: []string{}, Status: "unresolved"}
	}
	matches := false
	for _, ip := range ips {
		if serverIP != "" && net.ParseIP(ip) != nil && ip == serverIP {
			matches = true
		}
	}
	status := "resolved_other_ip"
	if matches {
		status = "matches_server"
	} else if len(ips) == 0 {
		status = "unresolved"
	}
	return Result{ResolvedIPs: ips, MatchesServer: matches, Status: status}
}
