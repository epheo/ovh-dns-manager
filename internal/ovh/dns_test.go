package ovh

import (
	"encoding/json"
	"testing"

	"ovh-dns-manager/internal/config"
)

func TestRecordsEqualDefaultTTL(t *testing.T) {
	desired := &config.DNSRecord{Name: "www", Type: "CNAME", Target: "home.example.com."}
	live := &config.DNSRecord{Name: "www", Type: "CNAME", Target: "home.example.com.", TTL: config.DefaultTTL}
	if !RecordsEqual(desired, live) {
		t.Fatal("unset TTL must match the default it is written with")
	}
	live.TTL = 0
	if !RecordsEqual(desired, live) {
		t.Fatal("unset TTL must match a zone-default live TTL")
	}
	live.TTL = 300
	if RecordsEqual(desired, live) {
		t.Fatal("unset TTL must not match a non-default live TTL")
	}
}

func TestCreateBodyHasNoPriority(t *testing.T) {
	body, err := json.Marshal(ConvertDNSRecordToOVHCreate(&config.DNSRecord{Type: "MX", Target: "10 home.example.com."}))
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]any
	if err := json.Unmarshal(body, &fields); err != nil {
		t.Fatal(err)
	}
	if _, ok := fields["priority"]; ok {
		t.Fatalf("body %s has priority, which domain.zone.RecordCreate lacks", body)
	}
	if fields["target"] != "10 home.example.com." {
		t.Fatalf("target = %v", fields["target"])
	}
}
