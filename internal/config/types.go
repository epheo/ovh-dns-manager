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
	Priority int    `yaml:"priority,omitempty"`
}

type OVHRecord struct {
	ID       int64  `json:"id,omitempty"`
	Zone     string `json:"zone"`
	SubDomain string `json:"subDomain"`
	FieldType string `json:"fieldType"`
	Target   string `json:"target"`
	TTL      int    `json:"ttl"`
	Priority *int   `json:"priority,omitempty"`
}

type OVHRecordCreate struct {
	SubDomain string `json:"subDomain"`
	FieldType string `json:"fieldType"`
	Target   string `json:"target"`
	TTL      int    `json:"ttl"`
	Priority *int   `json:"priority,omitempty"`
}

type OVHRecordUpdate struct {
	Target   string `json:"target"`
	TTL      int    `json:"ttl"`
	Priority *int   `json:"priority,omitempty"`
}

type OVHCredentials struct {
	Endpoint         string `yaml:"endpoint"`
	ApplicationKey   string `yaml:"application_key"`
	ApplicationSecret string `yaml:"application_secret"`
	ConsumerKey      string `yaml:"consumer_key"`
	Timeout          int    `yaml:"timeout"`
}