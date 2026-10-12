package ovh

import (
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
