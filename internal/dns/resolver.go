package dns

import (
	"context"
	"net"
	"sort"
)

type Resolver struct{ client *net.Resolver }

func NewResolver() *Resolver { return &Resolver{client: net.DefaultResolver} }
func (r *Resolver) LookupIP(ctx context.Context, host string) ([]string, error) {
	ips, err := r.client.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}
	unique := map[string]bool{}
	var result []string
	for _, ip := range ips {
		s := ip.IP.String()
		if !unique[s] {
			unique[s] = true
			result = append(result, s)
		}
	}
	sort.Strings(result)
	return result, nil
}
