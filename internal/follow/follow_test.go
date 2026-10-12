package follow

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ovh-dns-manager/internal/config"
)

func doh(t *testing.T, ips ...string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var answers []string
		answers = append(answers, `{"type":5,"data":"dyn.example.net."}`)
		for _, ip := range ips {
			answers = append(answers, fmt.Sprintf(`{"type":1,"data":%q}`, ip))
		}
		fmt.Fprintf(w, `{"Status":0,"Answer":[%s]}`, strings.Join(answers, ","))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func resolver(srvs ...*httptest.Server) *Resolver {
	r := NewResolver()
	r.Endpoints = nil
	for _, s := range srvs {
		r.Endpoints = append(r.Endpoints, s.URL)
	}
	return r
}

func TestResolve(t *testing.T) {
	cases := []struct {
		name    string
		a, b    []string
		want    string
		wantErr string
	}{
		{"agree", []string{"203.0.113.7"}, []string{"203.0.113.7"}, "203.0.113.7", ""},
		{"disagree", []string{"203.0.113.7"}, []string{"203.0.113.8"}, "", "disagree"},
		{"empty", nil, nil, "", "exactly one"},
		{"several", []string{"203.0.113.7", "203.0.113.8"}, []string{"203.0.113.7", "203.0.113.8"}, "", "exactly one"},
		{"private", []string{"192.168.1.1"}, []string{"192.168.1.1"}, "", "not a public"},
		{"cgnat", []string{"100.64.0.1"}, []string{"100.64.0.1"}, "", "not a public"},
		{"loopback", []string{"127.0.0.1"}, []string{"127.0.0.1"}, "", "not a public"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := resolver(doh(t, c.a...), doh(t, c.b...)).Resolve("home.example.com")
			if c.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), c.wantErr) {
					t.Fatalf("err = %v, want %q", err, c.wantErr)
				}
				return
			}
			if err != nil || got != c.want {
				t.Fatalf("got %q, %v; want %q", got, err, c.want)
			}
		})
	}
}

func TestResolveZoneMarksUnresolved(t *testing.T) {
	zone := &config.DNSZone{Records: []config.DNSRecord{
		{Name: "", Type: "A", Follow: "home.example.com"},
		{Name: "www", Type: "A", Follow: "home.example.com"},
		{Name: "mail", Type: "A", Target: "198.51.100.1"},
	}}
	errs := ResolveZone(zone, resolver(doh(t, "203.0.113.7"), doh(t, "203.0.113.9")))
	// The cache holds only successes, so www retries and fails too.
	if len(errs) != 2 {
		t.Fatalf("errs = %v, want 2", errs)
	}
	if !zone.Records[0].Unresolved || !zone.Records[1].Unresolved || zone.Records[2].Unresolved {
		t.Fatalf("unexpected Unresolved flags: %+v", zone.Records)
	}
	if zone.Records[0].Target != "" {
		t.Fatalf("unresolved record got a target: %q", zone.Records[0].Target)
	}
}

func TestResolveZoneFillsTarget(t *testing.T) {
	zone := &config.DNSZone{Records: []config.DNSRecord{
		{Name: "", Type: "A", Follow: "home.example.com"},
	}}
	if errs := ResolveZone(zone, resolver(doh(t, "203.0.113.7"), doh(t, "203.0.113.7"))); errs != nil {
		t.Fatal(errs)
	}
	if zone.Records[0].Target != "203.0.113.7" {
		t.Fatalf("target = %q", zone.Records[0].Target)
	}
}
