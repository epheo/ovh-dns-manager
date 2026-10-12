// Package follow resolves `follow:` records to the current public IPv4 of
// another host, so an apex A record can track a dynamic address.
package follow

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/netip"
	"slices"
	"time"

	"ovh-dns-manager/internal/config"
)

// Two independent DoH providers must agree: while a dynamic-DNS update is
// mid-flight they can disagree, and writing a stale apex is worse than
// keeping the current one for another run.
var DefaultEndpoints = []string{
	"https://cloudflare-dns.com/dns-query",
	"https://dns.google/resolve",
}

type Resolver struct {
	Endpoints []string
	Client    *http.Client
}

func NewResolver() *Resolver {
	return &Resolver{
		Endpoints: DefaultEndpoints,
		Client:    &http.Client{Timeout: 10 * time.Second},
	}
}

type dohResponse struct {
	Status int `json:"Status"`
	Answer []struct {
		Type int    `json:"type"`
		Data string `json:"data"`
	} `json:"Answer"`
}

func (r *Resolver) query(endpoint, host string) ([]string, error) {
	req, err := http.NewRequest("GET", endpoint+"?name="+host+"&type=A", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/dns-json")
	resp, err := r.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: HTTP %d", endpoint, resp.StatusCode)
	}
	var d dohResponse
	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		return nil, fmt.Errorf("%s: %w", endpoint, err)
	}
	if d.Status != 0 {
		return nil, fmt.Errorf("%s: DNS status %d", endpoint, d.Status)
	}
	var ips []string
	for _, a := range d.Answer {
		// CNAME hops (type 5) appear in the chain; only the final A counts.
		if a.Type == 1 {
			ips = append(ips, a.Data)
		}
	}
	slices.Sort(ips)
	return ips, nil
}

// Resolve returns host's single public IPv4, or an error if the providers
// disagree, return none or several, or return a non-public address.
func (r *Resolver) Resolve(host string) (string, error) {
	var first []string
	for i, ep := range r.Endpoints {
		ips, err := r.query(ep, host)
		if err != nil {
			return "", err
		}
		if i == 0 {
			first = ips
		} else if !slices.Equal(first, ips) {
			return "", fmt.Errorf("%s: resolvers disagree: %v vs %v", host, first, ips)
		}
	}
	if len(first) != 1 {
		return "", fmt.Errorf("%s: expected exactly one A record, got %v", host, first)
	}
	if err := checkPublic(first[0]); err != nil {
		return "", fmt.Errorf("%s: %w", host, err)
	}
	return first[0], nil
}

var cgnat = netip.MustParsePrefix("100.64.0.0/10")

func checkPublic(s string) error {
	ip, err := netip.ParseAddr(s)
	if err != nil || !ip.Is4() {
		return fmt.Errorf("%q is not an IPv4 address", s)
	}
	if !ip.IsGlobalUnicast() || ip.IsPrivate() || cgnat.Contains(ip) {
		return fmt.Errorf("%s is not a public address", s)
	}
	return nil
}

// ResolveZone fills Target on every follow record. A record that cannot be
// resolved is marked Unresolved so the sync leaves the live record alone.
func ResolveZone(zone *config.DNSZone, r *Resolver) []error {
	cache := map[string]string{}
	var errs []error
	for i := range zone.Records {
		rec := &zone.Records[i]
		if rec.Follow == "" {
			continue
		}
		ip, ok := cache[rec.Follow]
		if !ok {
			var err error
			ip, err = r.Resolve(rec.Follow)
			if err != nil {
				rec.Unresolved = true
				errs = append(errs, err)
				continue
			}
			cache[rec.Follow] = ip
		}
		rec.Target = ip
	}
	return errs
}
