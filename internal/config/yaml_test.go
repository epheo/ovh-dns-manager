package config

import "testing"

func TestValidateFollowAndExternal(t *testing.T) {
	cases := []struct {
		name string
		rec  DNSRecord
		ok   bool
	}{
		{"follow on A", DNSRecord{Type: "A", Follow: "home.example.com"}, true},
		{"follow on CNAME", DNSRecord{Type: "CNAME", Follow: "home.example.com"}, false},
		{"follow with target", DNSRecord{Type: "A", Follow: "home.example.com", Target: "203.0.113.7"}, false},
		{"external bare", DNSRecord{Type: "A", External: true}, true},
		{"external with target", DNSRecord{Type: "A", External: true, Target: "203.0.113.7"}, false},
		{"plain without target", DNSRecord{Type: "A"}, false},
	}
	for _, c := range cases {
		if err := ValidateDNSRecord(&c.rec); (err == nil) != c.ok {
			t.Errorf("%s: err = %v, want ok=%v", c.name, err, c.ok)
		}
	}
}

func TestValidatePriorityBelongsInTarget(t *testing.T) {
	if err := ValidateDNSRecord(&DNSRecord{Type: "MX", Target: "10 mail.example.com."}); err != nil {
		t.Fatalf("MX with its priority in the target: %v", err)
	}
	if err := ValidateDNSRecord(&DNSRecord{Type: "MX", Target: "mail.example.com.", Priority: 10}); err == nil {
		t.Fatal("a priority field must be refused: OVH records have none")
	}
}
