package discovery

import (
	"context"
	"domain-scout/internal/inspectors"
	"domain-scout/internal/ssh"
)

type Scanner struct{ Inspectors []inspectors.Inspector }

func (s *Scanner) Scan(ctx context.Context, ex ssh.Executor) ([]inspectors.Domain, []string) {
	var domains []inspectors.Domain
	var services []string
	for _, i := range s.Inspectors {
		ok, err := i.Detect(ctx, ex)
		if err != nil || !ok {
			continue
		}
		services = append(services, i.Name())
		found, err := i.Discover(ctx, ex)
		if err == nil {
			domains = append(domains, found...)
		}
	}
	return Normalize(domains), services
}
