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
	result := Result{
		ScanResponse: ScanResponse{
			Servers: []ServerInfo{},
			System:  SystemInfo{IPv4: []string{}, IPv6: []string{}},
		},
		Status: "completed", ScannedAt: time.Now().UTC(),
		DetectedServices: []string{}, Domains: []DomainResult{}, Warnings: []string{},
	}
	info, err := s.serverService.GetInfo(ctx, ex)
	if err != nil {
		return result, err
	}
	result.Server = info
	var found []inspectors.Domain
	for _, ins := range s.inspectors {
		if ctx.Err() != nil {
			return result, ctx.Err()
		}
		if collector, ok := ins.(interface {
			Collect(context.Context, ssh.Executor) (server.SystemInfo, []string, error)
		}); ok {
			systemInfo, warnings, collectErr := collector.Collect(ctx, ex)
			result.System = systemInfo
			result.Warnings = append(result.Warnings, warnings...)
			if collectErr != nil {
				if ctx.Err() != nil {
					return result, ctx.Err()
				}
				result.Warnings = append(result.Warnings, ins.Name()+" information: "+collectErr.Error())
			}
		}
		ok, detectErr := ins.Detect(ctx, ex)
		if detectErr != nil {
			if ctx.Err() != nil {
				return result, ctx.Err()
			}
			result.Warnings = append(result.Warnings, ins.Name()+" detection: "+detectErr.Error())
			continue
		}
		if !ok {
			continue
		}
		result.DetectedServices = append(result.DetectedServices, ins.Name())

		if ins.Name() == "system" {
			continue
		}

		serverInfo := ServerInfo{
			Server: ins.Name(), Domains: []DomainInfo{},
		}
		if statusInspector, ok := ins.(interface {
			Running(context.Context, ssh.Executor) (bool, error)
		}); ok {
			running, statusErr := statusInspector.Running(ctx, ex)
			if statusErr != nil {
				if ctx.Err() != nil {
					return result, ctx.Err()
				}
				result.Warnings = append(result.Warnings, ins.Name()+" status: "+statusErr.Error())
			} else {
				serverInfo.Running = running
			}
		}
		domains, scanErr := ins.Discover(ctx, ex)
		if scanErr != nil {
			if ctx.Err() != nil {
				return result, ctx.Err()
			}
			result.Warnings = append(result.Warnings, ins.Name()+" scan: "+scanErr.Error())
		}
		for _, domain := range domains {
			serverInfo.Domains = append(serverInfo.Domains, DomainInfo{
				Domain: domain.Name, Aliases: nonNilStrings(domain.Aliases),
				ConfigFile: domain.ConfigPath, DocumentRoot: domain.DocumentRoot,
				Ports: nonNilInts(domain.Ports),
			})
			found = append(found, domain)
			for _, alias := range domain.Aliases {
				legacyAlias := domain
				legacyAlias.Name = alias
				legacyAlias.Aliases = nil
				found = append(found, legacyAlias)
			}
		}
		result.Servers = append(result.Servers, serverInfo)
	}
	if ctx.Err() != nil {
		return result, ctx.Err()
	}
	found = Normalize(found)
	for _, d := range found {
		if ctx.Err() != nil {
			return result, ctx.Err()
		}
		dr := DomainResult{Domain: d}
		dr.DNS = s.verifier.Verify(ctx, d.Name, info.PrimaryIP)
		result.Domains = append(result.Domains, dr)
	}
	result.TotalDomains = len(result.Domains)
	return result, nil
}

func nonNilStrings(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}

func nonNilInts(values []int) []int {
	if values == nil {
		return []int{}
	}
	return values
}
