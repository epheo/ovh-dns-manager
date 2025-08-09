package ovh

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"ovh-dns-manager/internal/config"
)

func (c *Client) GetZoneRecords(zoneName string) ([]config.OVHRecord, error) {
	path := fmt.Sprintf("/domain/zone/%s/record", zoneName)
	resp, err := c.doRequest("GET", path, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get zone records: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var recordIDs []int64
	if err := json.Unmarshal(body, &recordIDs); err != nil {
		return nil, fmt.Errorf("failed to parse record IDs: %w", err)
	}

	var records []config.OVHRecord
	for _, id := range recordIDs {
		record, err := c.GetRecord(zoneName, id)
		if err != nil {
			return nil, fmt.Errorf("failed to get record %d: %w", id, err)
		}
		records = append(records, *record)
	}

	return records, nil
}

func (c *Client) GetRecord(zoneName string, recordID int64) (*config.OVHRecord, error) {
	path := fmt.Sprintf("/domain/zone/%s/record/%d", zoneName, recordID)
	resp, err := c.doRequest("GET", path, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get record: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var record config.OVHRecord
	if err := json.Unmarshal(body, &record); err != nil {
		return nil, fmt.Errorf("failed to parse record: %w", err)
	}

	return &record, nil
}

func (c *Client) CreateRecord(zoneName string, record *config.OVHRecordCreate) (*config.OVHRecord, error) {
	path := fmt.Sprintf("/domain/zone/%s/record", zoneName)
	
	body, err := json.Marshal(record)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal record: %w", err)
	}

	resp, err := c.doRequest("POST", path, string(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create record: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var createdRecord config.OVHRecord
	if err := json.Unmarshal(respBody, &createdRecord); err != nil {
		return nil, fmt.Errorf("failed to parse created record: %w", err)
	}

	return &createdRecord, nil
}

func (c *Client) UpdateRecord(zoneName string, recordID int64, record *config.OVHRecordUpdate) error {
	path := fmt.Sprintf("/domain/zone/%s/record/%d", zoneName, recordID)
	
	body, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("failed to marshal record: %w", err)
	}

	resp, err := c.doRequest("PUT", path, string(body))
	if err != nil {
		return fmt.Errorf("failed to update record: %w", err)
	}
	resp.Body.Close()

	return nil
}

func (c *Client) DeleteRecord(zoneName string, recordID int64) error {
	path := fmt.Sprintf("/domain/zone/%s/record/%d", zoneName, recordID)
	
	resp, err := c.doRequest("DELETE", path, "")
	if err != nil {
		return fmt.Errorf("failed to delete record: %w", err)
	}
	resp.Body.Close()

	return nil
}

func (c *Client) RefreshZone(zoneName string) error {
	path := fmt.Sprintf("/domain/zone/%s/refresh", zoneName)
	
	resp, err := c.doRequest("POST", path, "")
	if err != nil {
		return fmt.Errorf("failed to refresh zone: %w", err)
	}
	resp.Body.Close()

	return nil
}

func (c *Client) GetZones() ([]string, error) {
	path := "/domain/zone"
	resp, err := c.doRequest("GET", path, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get zones: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var zones []string
	if err := json.Unmarshal(body, &zones); err != nil {
		return nil, fmt.Errorf("failed to parse zones: %w", err)
	}

	return zones, nil
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

func ConvertDNSRecordToOVHCreate(dnsRecord *config.DNSRecord) *config.OVHRecordCreate {
	record := &config.OVHRecordCreate{
		SubDomain: dnsRecord.Name,
		FieldType: dnsRecord.Type,
		Target:    dnsRecord.Target,
		TTL:       dnsRecord.TTL,
	}

	if dnsRecord.Priority > 0 {
		record.Priority = &dnsRecord.Priority
	}

	if record.TTL == 0 {
		record.TTL = 3600
	}

	return record
}

func ConvertDNSRecordToOVHUpdate(dnsRecord *config.DNSRecord) *config.OVHRecordUpdate {
	record := &config.OVHRecordUpdate{
		Target: dnsRecord.Target,
		TTL:    dnsRecord.TTL,
	}

	if dnsRecord.Priority > 0 {
		record.Priority = &dnsRecord.Priority
	}

	if record.TTL == 0 {
		record.TTL = 3600
	}

	return record
}

func RecordsEqual(a, b *config.DNSRecord) bool {
	return a.Name == b.Name &&
		a.Type == b.Type &&
		a.Target == b.Target &&
		a.TTL == b.TTL &&
		a.Priority == b.Priority
}

func FindRecordByKey(records []config.DNSRecord, name, recordType string) *config.DNSRecord {
	for _, record := range records {
		if record.Name == name && record.Type == recordType {
			return &record
		}
	}
	return nil
}

func RecordKey(record *config.DNSRecord) string {
	return record.Name + ":" + record.Type
}

func OVHRecordKey(record *config.OVHRecord) string {
	return record.SubDomain + ":" + record.FieldType
}

func RecordKeyWithID(record *config.OVHRecord) string {
	return record.SubDomain + ":" + record.FieldType + ":" + strconv.FormatInt(record.ID, 10)
}