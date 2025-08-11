package ovh

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"ovh-dns-manager/internal/config"
)

// readJSONResponse is a helper function to read and unmarshal JSON responses
func readJSONResponse(resp *http.Response, v interface{}) error {
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}
	
	if err := json.Unmarshal(body, v); err != nil {
		return fmt.Errorf("failed to parse JSON response: %w", err)
	}
	
	return nil
}

func (c *Client) GetZoneRecords(zoneName string) ([]config.OVHRecord, error) {
	path := fmt.Sprintf("/domain/zone/%s/record", zoneName)
	resp, err := c.doRequest("GET", path, "")
	if err != nil {
		return nil, err
	}

	var recordIDs []int64
	if err := readJSONResponse(resp, &recordIDs); err != nil {
		return nil, err
	}

	var records []config.OVHRecord
	for _, id := range recordIDs {
		record, err := c.GetRecord(zoneName, id)
		if err != nil {
			return nil, err
		}
		records = append(records, *record)
	}

	return records, nil
}

func (c *Client) GetRecord(zoneName string, recordID int64) (*config.OVHRecord, error) {
	path := fmt.Sprintf("/domain/zone/%s/record/%d", zoneName, recordID)
	resp, err := c.doRequest("GET", path, "")
	if err != nil {
		return nil, err
	}

	var record config.OVHRecord
	if err := readJSONResponse(resp, &record); err != nil {
		return nil, err
	}

	return &record, nil
}

func (c *Client) CreateRecord(zoneName string, record *config.OVHRecordCreate) (*config.OVHRecord, error) {
	path := fmt.Sprintf("/domain/zone/%s/record", zoneName)
	
	body, err := json.Marshal(record)
	if err != nil {
		return nil, err
	}

	resp, err := c.doRequest("POST", path, string(body))
	if err != nil {
		return nil, err
	}

	var createdRecord config.OVHRecord
	if err := readJSONResponse(resp, &createdRecord); err != nil {
		return nil, err
	}

	return &createdRecord, nil
}

func (c *Client) UpdateRecord(zoneName string, recordID int64, record *config.OVHRecordUpdate) error {
	path := fmt.Sprintf("/domain/zone/%s/record/%d", zoneName, recordID)
	
	body, err := json.Marshal(record)
	if err != nil {
		return err
	}

	resp, err := c.doRequest("PUT", path, string(body))
	if err != nil {
		return err
	}
	resp.Body.Close()

	return nil
}

func (c *Client) DeleteRecord(zoneName string, recordID int64) error {
	path := fmt.Sprintf("/domain/zone/%s/record/%d", zoneName, recordID)
	
	resp, err := c.doRequest("DELETE", path, "")
	if err != nil {
		return err
	}
	resp.Body.Close()

	return nil
}

func (c *Client) RefreshZone(zoneName string) error {
	path := fmt.Sprintf("/domain/zone/%s/refresh", zoneName)
	
	resp, err := c.doRequest("POST", path, "")
	if err != nil {
		return err
	}
	resp.Body.Close()

	return nil
}

func ConvertOVHRecordToDNSRecord(ovhRecord *config.OVHRecord) *config.DNSRecord {
	record := &config.DNSRecord{
		Name:   ovhRecord.SubDomain,
		Type:   ovhRecord.FieldType,
		Target: ovhRecord.Target,
		TTL:    ovhRecord.TTL,
	}

	if ovhRecord.Priority != nil {
		record.Priority = *ovhRecord.Priority
	}

	return record
}

// setRecordDefaults applies default TTL and priority handling
func setRecordDefaults(ttl int, priority int, recordType string) (int, *int) {
	if ttl == 0 {
		ttl = config.DefaultTTL
	}
	
	var priorityPtr *int
	// Only include priority for record types that support it
	if (recordType == "MX" || recordType == "SRV") && priority >= 0 {
		priorityPtr = &priority
	}
	
	return ttl, priorityPtr
}

func ConvertDNSRecordToOVHCreate(dnsRecord *config.DNSRecord) *config.OVHRecordCreate {
	ttl, priority := setRecordDefaults(dnsRecord.TTL, dnsRecord.Priority, dnsRecord.Type)
	
	return &config.OVHRecordCreate{
		SubDomain: dnsRecord.Name,
		FieldType: dnsRecord.Type,
		Target:    dnsRecord.Target,
		TTL:       ttl,
		Priority:  priority,
	}
}

func ConvertDNSRecordToOVHUpdate(dnsRecord *config.DNSRecord) *config.OVHRecordUpdate {
	ttl, priority := setRecordDefaults(dnsRecord.TTL, dnsRecord.Priority, dnsRecord.Type)
	
	return &config.OVHRecordUpdate{
		Target:   dnsRecord.Target,
		TTL:      ttl,
		Priority: priority,
	}
}

func RecordsEqual(a, b *config.DNSRecord) bool {
	return a.Name == b.Name &&
		a.Type == b.Type &&
		a.Target == b.Target &&
		a.TTL == b.TTL &&
		a.Priority == b.Priority
}

func RecordKey(record *config.DNSRecord) string {
	return record.Name + ":" + record.Type
}

func OVHRecordKey(record *config.OVHRecord) string {
	return record.SubDomain + ":" + record.FieldType
}

