package discovery

import (
	"context"
	"domain-scout/internal/dns"
	"domain-scout/internal/inspectors"
	"domain-scout/internal/server"
	"domain-scout/internal/ssh"
	"time"
)

type Service struct {
	inspectors    []inspectors.Inspector
	serverService *server.Service
	verifier      *dns.Verifier
}

func NewService(list []inspectors.Inspector) *Service {
	return &Service{inspectors: list, serverService: server.NewService(), verifier: dns.NewVerifier()}
}
func (s *Service) Discover(ctx context.Context, ex ssh.Executor) (Result, error) {
	result := Result{Status: "completed", ScannedAt: time.Now().UTC(), DetectedServices: []string{}, Domains: []DomainResult{}, Warnings: []string{}}
	info, err := s.serverService.GetInfo(ctx, ex)
	if err != nil {
		return result, err
	}
	result.Server = info
	var found []inspectors.Domain
	for _, ins := range s.inspectors {
		ok, detectErr := ins.Detect(ctx, ex)
		if detectErr != nil {
			result.Warnings = append(result.Warnings, ins.Name()+" detection: "+detectErr.Error())
			continue
		}
		if !ok {
			continue
		}
		result.DetectedServices = append(result.DetectedServices, ins.Name())
		domains, scanErr := ins.Discover(ctx, ex)
		if scanErr != nil {
			result.Warnings = append(result.Warnings, ins.Name()+" scan: "+scanErr.Error())
			continue
		}
		found = append(found, domains...)
	}
	found = Normalize(found)
	for _, d := range found {
		dr := DomainResult{Domain: d}
		dr.DNS = s.verifier.Verify(ctx, d.Name, info.PrimaryIP)
		result.Domains = append(result.Domains, dr)
	}
	result.TotalDomains = len(result.Domains)
	return result, nil
}
