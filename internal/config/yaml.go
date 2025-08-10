package config

import (
	"fmt"
	"os"
	"strconv"

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
	var creds OVHCredentials
	
	// First, try to load from environment variables
	creds.Endpoint = getEnvOrDefault("OVH_ENDPOINT", "")
	creds.ApplicationKey = getEnvOrDefault("OVH_APPLICATION_KEY", "")
	creds.ApplicationSecret = getEnvOrDefault("OVH_APPLICATION_SECRET", "")
	creds.ConsumerKey = getEnvOrDefault("OVH_CONSUMER_KEY", "")
	creds.Timeout = getEnvIntOrDefault("OVH_TIMEOUT", 0)

	// Check if we have all required credentials from environment
	hasEnvCreds := creds.ApplicationKey != "" && creds.ApplicationSecret != "" && creds.ConsumerKey != ""
	
	// If not all credentials from env, try to load from file
	if !hasEnvCreds {
		if filename == "" {
			filename = "ovh-credentials.yaml"
		}

		data, err := os.ReadFile(filename)
		if err != nil {
			if hasEnvCreds {
				// If we have some env vars but file failed, continue with env vars only
			} else {
				return nil, fmt.Errorf("failed to read credentials file %s and no environment variables set: %w", filename, err)
			}
		} else {
			var fileCreds OVHCredentials
			if err := yaml.Unmarshal(data, &fileCreds); err != nil {
				return nil, fmt.Errorf("failed to parse credentials YAML: %w", err)
			}
			
			// Use file values only if env vars are not set
			if creds.Endpoint == "" {
				creds.Endpoint = fileCreds.Endpoint
			}
			if creds.ApplicationKey == "" {
				creds.ApplicationKey = fileCreds.ApplicationKey
			}
			if creds.ApplicationSecret == "" {
				creds.ApplicationSecret = fileCreds.ApplicationSecret
			}
			if creds.ConsumerKey == "" {
				creds.ConsumerKey = fileCreds.ConsumerKey
			}
			if creds.Timeout == 0 {
				creds.Timeout = fileCreds.Timeout
			}
		}
	}

	// Apply defaults
	if creds.Endpoint == "" {
		creds.Endpoint = "ovh-eu"
	}
	if creds.Timeout == 0 {
		creds.Timeout = 30
	}

	// Validate required fields
	if creds.ApplicationKey == "" {
		return nil, fmt.Errorf("application_key is required (set OVH_APPLICATION_KEY env var or provide in credentials file)")
	}
	if creds.ApplicationSecret == "" {
		return nil, fmt.Errorf("application_secret is required (set OVH_APPLICATION_SECRET env var or provide in credentials file)")
	}
	if creds.ConsumerKey == "" {
		return nil, fmt.Errorf("consumer_key is required (set OVH_CONSUMER_KEY env var or provide in credentials file)")
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
	case "A", "AAAA", "CNAME", "TXT", "NS", "SPF", "CAA", "PTR":
		// These types don't require priority
	case "MX", "SRV":
		// These types require priority (0 is valid for MX)
		// Priority validation is handled elsewhere if needed
	default:
		return fmt.Errorf("unsupported record type: %s", record.Type)
	}

	if record.TTL < 0 {
		return fmt.Errorf("TTL cannot be negative")
	}

	return nil
}

// getEnvOrDefault returns the environment variable value or the default if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvIntOrDefault returns the environment variable as int or the default if not set or invalid
func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}