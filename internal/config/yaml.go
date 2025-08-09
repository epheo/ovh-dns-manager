package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func LoadDNSZone(filename string) (*DNSZone, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	var zone DNSZone
	if err := yaml.Unmarshal(data, &zone); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &zone, nil
}

func SaveDNSZone(zone *DNSZone, filename string) error {
	data, err := yaml.Marshal(zone)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", filename, err)
	}

	return nil
}

func LoadOVHCredentials(filename string) (*OVHCredentials, error) {
	if filename == "" {
		filename = "ovh-credentials.yaml"
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read credentials file %s: %w", filename, err)
	}

	var creds OVHCredentials
	if err := yaml.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("failed to parse credentials YAML: %w", err)
	}

	if creds.Endpoint == "" {
		creds.Endpoint = "ovh-eu"
	}
	if creds.Timeout == 0 {
		creds.Timeout = 30
	}

	return &creds, nil
}

func ValidateDNSRecord(record *DNSRecord) error {
	if record.Name == "" && record.Type != "A" && record.Type != "AAAA" && record.Type != "MX" && record.Type != "TXT" {
		// Allow empty name only for root domain records
	}

	if record.Type == "" {
		return fmt.Errorf("record type is required")
	}

	if record.Target == "" {
		return fmt.Errorf("record target is required")
	}

	switch record.Type {
	case "A", "AAAA", "CNAME", "TXT", "NS":
		// These types don't require priority
	case "MX", "SRV":
		// These types require priority
		if record.Priority == 0 {
			return fmt.Errorf("priority is required for %s records", record.Type)
		}
	default:
		return fmt.Errorf("unsupported record type: %s", record.Type)
	}

	if record.TTL < 0 {
		return fmt.Errorf("TTL cannot be negative")
	}

	return nil
}