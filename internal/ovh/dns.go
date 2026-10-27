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
	return &config.DNSRecord{
		Name:   ovhRecord.SubDomain,
		Type:   ovhRecord.FieldType,
		Target: ovhRecord.Target,
		TTL:    ovhRecord.TTL,
	}
}

func defaultTTL(ttl int) int {
	if ttl == 0 {
		return config.DefaultTTL
	}
	return ttl
}

func ConvertDNSRecordToOVHCreate(dnsRecord *config.DNSRecord) *config.OVHRecordCreate {
	return &config.OVHRecordCreate{
		SubDomain: dnsRecord.Name,
		FieldType: dnsRecord.Type,
		Target:    dnsRecord.Target,
		TTL:       defaultTTL(dnsRecord.TTL),
	}
}

func ConvertDNSRecordToOVHUpdate(dnsRecord *config.DNSRecord) *config.OVHRecordUpdate {
	return &config.OVHRecordUpdate{
		Target: dnsRecord.Target,
		TTL:    defaultTTL(dnsRecord.TTL),
	}
}

// RecordsEqual compares a desired record with a live one. An unset desired
// TTL accepts 0 (zone default, records made in the OVH UI) and DefaultTTL
// (what create and update send); comparing it literally rewrote every such
// record on every run.
func RecordsEqual(desired, current *config.DNSRecord) bool {
	ttlOK := desired.TTL == current.TTL
	if desired.TTL == 0 {
		ttlOK = current.TTL == 0 || current.TTL == config.DefaultTTL
	}
	return desired.Name == current.Name &&
		desired.Type == current.Type &&
		desired.Target == current.Target &&
		ttlOK
}

func RecordKey(record *config.DNSRecord) string {
	return record.Name + ":" + record.Type
}

func OVHRecordKey(record *config.OVHRecord) string {
	return record.SubDomain + ":" + record.FieldType
}

