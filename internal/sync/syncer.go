package sync

import (
	"fmt"
	"log"

	"ovh-dns-manager/internal/config"
	"ovh-dns-manager/internal/ovh"
)

type Syncer struct {
	client *ovh.Client
	dryRun bool
}

type SyncResult struct {
	Created []config.DNSRecord
	Updated []config.DNSRecord
	Deleted []config.OVHRecord
	Errors  []error
}

func NewSyncer(client *ovh.Client, dryRun bool) *Syncer {
	return &Syncer{
		client: client,
		dryRun: dryRun,
	}
}

func (s *Syncer) SyncZone(zone *config.DNSZone) (*SyncResult, error) {
	result := &SyncResult{}

	currentRecords, err := s.client.GetZoneRecords(zone.Domain)
	if err != nil {
		return result, err
	}

	desiredRecords := make(map[string]*config.DNSRecord)
	for i := range zone.Records {
		key := ovh.RecordKey(&zone.Records[i])
		desiredRecords[key] = &zone.Records[i]
	}

	currentRecordsMap := make(map[string]*config.OVHRecord)
	for i := range currentRecords {
		key := ovh.OVHRecordKey(&currentRecords[i])
		currentRecordsMap[key] = &currentRecords[i]
	}

	for key, desired := range desiredRecords {
		current, exists := currentRecordsMap[key]
		if !exists {
			log.Printf("Creating record: %s %s -> %s", desired.Name, desired.Type, desired.Target)
			if !s.dryRun {
				createRecord := ovh.ConvertDNSRecordToOVHCreate(desired)
				_, err := s.client.CreateRecord(zone.Domain, createRecord)
				if err != nil {
					result.Errors = append(result.Errors, fmt.Errorf("failed to create record %s: %w", key, err))
					continue
				}
			}
			result.Created = append(result.Created, *desired)
		} else {
			currentDNS := ovh.ConvertOVHRecordToDNSRecord(current)
			if !ovh.RecordsEqual(desired, currentDNS) {
				log.Printf("Updating record: %s %s -> %s (was %s)", desired.Name, desired.Type, desired.Target, currentDNS.Target)
				if !s.dryRun {
					updateRecord := ovh.ConvertDNSRecordToOVHUpdate(desired)
					err := s.client.UpdateRecord(zone.Domain, current.ID, updateRecord)
					if err != nil {
						result.Errors = append(result.Errors, fmt.Errorf("failed to update record %s: %w", key, err))
						continue
					}
				}
				result.Updated = append(result.Updated, *desired)
			}
		}
	}

	for key, current := range currentRecordsMap {
		if _, exists := desiredRecords[key]; !exists {
			log.Printf("Deleting record: %s %s -> %s (ID: %d)", current.SubDomain, current.FieldType, current.Target, current.ID)
			if !s.dryRun {
				err := s.client.DeleteRecord(zone.Domain, current.ID)
				if err != nil {
					result.Errors = append(result.Errors, fmt.Errorf("failed to delete record %s: %w", key, err))
					continue
				}
			}
			result.Deleted = append(result.Deleted, *current)
		}
	}

	if result.HasChanges() && !s.dryRun {
		log.Printf("Refreshing DNS zone %s", zone.Domain)
		if err := s.client.RefreshZone(zone.Domain); err != nil {
			result.Errors = append(result.Errors, err)
		}
	}

	return result, nil
}

func (s *Syncer) ExportZone(domain string) (*config.DNSZone, error) {
	records, err := s.client.GetZoneRecords(domain)
	if err != nil {
		return nil, err
	}

	zone := &config.DNSZone{
		Domain:  domain,
		Records: make([]config.DNSRecord, 0, len(records)),
	}

	for _, ovhRecord := range records {
		dnsRecord := ovh.ConvertOVHRecordToDNSRecord(&ovhRecord)
		zone.Records = append(zone.Records, *dnsRecord)
	}

	return zone, nil
}

func (r *SyncResult) HasChanges() bool {
	return len(r.Created)+len(r.Updated)+len(r.Deleted) > 0
}

func (r *SyncResult) HasErrors() bool {
	return len(r.Errors) > 0
}

func (r *SyncResult) PrintSummary() {
	if !r.HasChanges() {
		log.Println("No changes needed")
		return
	}

	log.Printf("Summary: %d created, %d updated, %d deleted",
		len(r.Created), len(r.Updated), len(r.Deleted))

	if r.HasErrors() {
		log.Printf("Errors: %d", len(r.Errors))
		for _, err := range r.Errors {
			log.Printf("  - %v", err)
		}
	}
}

