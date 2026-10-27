package config

type DNSZone struct {
	Domain  string      `yaml:"domain"`
	Records []DNSRecord `yaml:"records"`
}

type DNSRecord struct {
	Name     string `yaml:"name"`
	Type     string `yaml:"type"`
	Target   string `yaml:"target"`
	TTL      int    `yaml:"ttl,omitempty"`
	// Priority is parsed only to be refused: OVH records have no such field,
	// MX and SRV carry it at the start of Target ("10 mail.example.com.").
	Priority int `yaml:"priority,omitempty"`
	// Follow names a host whose public IPv4 becomes Target at apply time,
	// for apexes that cannot be a CNAME to a dynamic address.
	Follow string `yaml:"follow,omitempty"`
	// External declares a record another writer owns (a router's DynHost):
	// apply never creates, updates or deletes it.
	External bool `yaml:"external,omitempty"`
	// Unresolved marks a Follow record whose address could not be trusted
	// this run; the syncer leaves the live record untouched.
	Unresolved bool `yaml:"-"`
}

type OVHRecord struct {
	ID       int64  `json:"id,omitempty"`
	Zone     string `json:"zone"`
	SubDomain string `json:"subDomain"`
	FieldType string `json:"fieldType"`
	Target   string `json:"target"`
	TTL      int    `json:"ttl"`
}

// The bodies follow OVH's domain.zone.RecordCreate and domain.zone.Record
// schemas, which have no priority field.
type OVHRecordCreate struct {
	SubDomain string `json:"subDomain"`
	FieldType string `json:"fieldType"`
	Target   string `json:"target"`
	TTL      int    `json:"ttl"`
}

type OVHRecordUpdate struct {
	Target   string `json:"target"`
	TTL      int    `json:"ttl"`
}

type OVHCredentials struct {
	Endpoint         string `yaml:"endpoint"`
	ApplicationKey   string `yaml:"application_key"`
	ApplicationSecret string `yaml:"application_secret"`
	ConsumerKey      string `yaml:"consumer_key"`
	Timeout          int    `yaml:"timeout"`
}