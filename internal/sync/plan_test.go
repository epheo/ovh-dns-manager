package sync

import (
	"testing"

	"ovh-dns-manager/internal/config"
)

func live(id int64, name, typ, target string, ttl int) config.OVHRecord {
	return config.OVHRecord{ID: id, SubDomain: name, FieldType: typ, Target: target, TTL: ttl}
}

func TestPlanMultiValueSetUnchanged(t *testing.T) {
	desired := []config.DNSRecord{
		{Type: "NS", Target: "dns1.example.net."},
		{Type: "NS", Target: "ns1.example.net."},
	}
	current := []config.OVHRecord{
		live(1, "", "NS", "ns1.example.net.", 0),
		live(2, "", "NS", "dns1.example.net.", 0),
	}
	p := Plan(desired, current)
	if len(p.Creates)+len(p.Updates)+len(p.Deletes) != 0 {
		t.Fatalf("expected no changes, got %+v", p)
	}
}

func TestPlanMultiValueSetAddAndRemove(t *testing.T) {
	desired := []config.DNSRecord{
		{Type: "TXT", Target: "keep"},
		{Type: "TXT", Target: "new"},
		{Type: "TXT", Target: "newer"},
	}
	current := []config.OVHRecord{
		live(1, "", "TXT", "keep", 0),
		live(2, "", "TXT", "old", 0),
	}
	p := Plan(desired, current)
	if len(p.Updates) != 1 || p.Updates[0].Current.ID != 2 || p.Updates[0].Desired.Target != "new" {
		t.Fatalf("updates = %+v", p.Updates)
	}
	if len(p.Creates) != 1 || p.Creates[0].Target != "newer" || len(p.Deletes) != 0 {
		t.Fatalf("creates = %+v deletes = %+v", p.Creates, p.Deletes)
	}
}

func TestPlanSingleValueRetargetsInPlace(t *testing.T) {
	p := Plan(
		[]config.DNSRecord{{Type: "A", Target: "203.0.113.8", TTL: 300}},
		[]config.OVHRecord{live(7, "", "A", "203.0.113.7", 300)},
	)
	if len(p.Updates) != 1 || p.Updates[0].Current.ID != 7 || len(p.Creates)+len(p.Deletes) != 0 {
		t.Fatalf("plan = %+v", p)
	}
}

func TestPlanDeletesUnlistedSet(t *testing.T) {
	p := Plan(nil, []config.OVHRecord{live(3, "old", "A", "203.0.113.7", 0)})
	if len(p.Deletes) != 1 || p.Deletes[0].ID != 3 {
		t.Fatalf("plan = %+v", p)
	}
}

func TestPlanSkipsExternalAndUnresolved(t *testing.T) {
	desired := []config.DNSRecord{
		{Name: "home", Type: "A", External: true},
		{Name: "", Type: "A", Follow: "home.example.com", Unresolved: true},
	}
	current := []config.OVHRecord{
		live(1, "home", "A", "203.0.113.7", 60),
		live(2, "", "A", "203.0.113.6", 300),
	}
	p := Plan(desired, current)
	if len(p.Creates)+len(p.Updates)+len(p.Deletes) != 0 || len(p.Skipped) != 2 {
		t.Fatalf("plan = %+v", p)
	}
}

func TestPlanTTLChangeUpdates(t *testing.T) {
	p := Plan(
		[]config.DNSRecord{{Name: "www", Type: "CNAME", Target: "home.example.com.", TTL: 300}},
		[]config.OVHRecord{live(5, "www", "CNAME", "home.example.com.", 3600)},
	)
	if len(p.Updates) != 1 || p.Updates[0].Current.ID != 5 {
		t.Fatalf("plan = %+v", p)
	}
}
