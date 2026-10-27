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

	plan := Plan(zone.Records, currentRecords)
	for _, sk := range plan.Skipped {
		log.Printf("Skipping %s: %s", sk.Key, sk.Reason)
	}

	for _, desired := range plan.Creates {
		log.Printf("Creating record: %s %s -> %s", desired.Name, desired.Type, desired.Target)
		if !s.dryRun {
			if _, err := s.client.CreateRecord(zone.Domain, ovh.ConvertDNSRecordToOVHCreate(&desired)); err != nil {
				result.Errors = append(result.Errors, fmt.Errorf("failed to create record %s: %w", ovh.RecordKey(&desired), err))
				continue
			}
		}
		result.Created = append(result.Created, desired)
	}

	for _, u := range plan.Updates {
		log.Printf("Updating record: %s %s -> %s (was %s)", u.Desired.Name, u.Desired.Type, u.Desired.Target, u.Current.Target)
		if !s.dryRun {
			if err := s.client.UpdateRecord(zone.Domain, u.Current.ID, ovh.ConvertDNSRecordToOVHUpdate(&u.Desired)); err != nil {
				result.Errors = append(result.Errors, fmt.Errorf("failed to update record %s: %w", ovh.RecordKey(&u.Desired), err))
				continue
			}
		}
		result.Updated = append(result.Updated, u.Desired)
	}

	for _, current := range plan.Deletes {
		log.Printf("Deleting record: %s %s -> %s (ID: %d)", current.SubDomain, current.FieldType, current.Target, current.ID)
		if !s.dryRun {
			if err := s.client.DeleteRecord(zone.Domain, current.ID); err != nil {
				result.Errors = append(result.Errors, fmt.Errorf("failed to delete record %s: %w", ovh.OVHRecordKey(&current), err))
				continue
			}
		}
		result.Deleted = append(result.Deleted, current)
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
	if r.HasChanges() {
		log.Printf("Summary: %d created, %d updated, %d deleted",
			len(r.Created), len(r.Updated), len(r.Deleted))
	} else {
		log.Println("No changes needed")
	}

	// Failed operations are not changes: a run where every one failed still
	// has to say why.
	if r.HasErrors() {
		log.Printf("Errors: %d", len(r.Errors))
		for _, err := range r.Errors {
			log.Printf("  - %v", err)
		}
	}
}

