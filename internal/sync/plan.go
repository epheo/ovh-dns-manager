package sync

import (
	"sort"

	"ovh-dns-manager/internal/config"
	"ovh-dns-manager/internal/ovh"
)

type Update struct {
	Desired config.DNSRecord
	Current config.OVHRecord
}

type Skip struct {
	Key    string
	Reason string
}

type ZonePlan struct {
	Creates []config.DNSRecord
	Updates []Update
	Deletes []config.OVHRecord
	Skipped []Skip
}

// Plan diffs desired against live records one record set (name and type) at
// a time. Keying single records by name and type collapsed multi-value sets
// such as the two apex NS records into one, so the others were rewritten or
// never reconciled.
func Plan(desired []config.DNSRecord, current []config.OVHRecord) ZonePlan {
	var plan ZonePlan

	want := map[string][]config.DNSRecord{}
	var order []string
	for _, d := range desired {
		k := ovh.RecordKey(&d)
		if _, ok := want[k]; !ok {
			order = append(order, k)
		}
		want[k] = append(want[k], d)
	}
	have := map[string][]config.OVHRecord{}
	for _, c := range current {
		k := ovh.OVHRecordKey(&c)
		if _, ok := want[k]; !ok && len(have[k]) == 0 {
			order = append(order, k)
		}
		have[k] = append(have[k], c)
	}

	for _, k := range order {
		ds, cs := want[k], have[k]
		if reason := skipReason(ds); reason != "" {
			plan.Skipped = append(plan.Skipped, Skip{Key: k, Reason: reason})
			continue
		}
		planSet(&plan, dedupe(ds), cs)
	}
	return plan
}

func skipReason(ds []config.DNSRecord) string {
	for _, d := range ds {
		if d.External {
			return "external, owned by another writer"
		}
		if d.Unresolved {
			return "follow " + d.Follow + " unresolved"
		}
	}
	return ""
}

func dedupe(ds []config.DNSRecord) []config.DNSRecord {
	seen := map[string]bool{}
	var out []config.DNSRecord
	for _, d := range ds {
		if !seen[d.Target] {
			seen[d.Target] = true
			out = append(out, d)
		}
	}
	return out
}

func planSet(plan *ZonePlan, ds []config.DNSRecord, cs []config.OVHRecord) {
	matched := map[int64]bool{}
	var leftD []config.DNSRecord
	for _, d := range ds {
		found := false
		for _, c := range cs {
			if matched[c.ID] || c.Target != d.Target {
				continue
			}
			matched[c.ID], found = true, true
			if !ovh.RecordsEqual(&d, ovh.ConvertOVHRecordToDNSRecord(&c)) {
				plan.Updates = append(plan.Updates, Update{Desired: d, Current: c})
			}
			break
		}
		if !found {
			leftD = append(leftD, d)
		}
	}

	var leftC []config.OVHRecord
	for _, c := range cs {
		if !matched[c.ID] {
			leftC = append(leftC, c)
		}
	}

	// Retargeting in place keeps a single-value record present throughout,
	// instead of a delete and create.
	sort.Slice(leftD, func(i, j int) bool { return leftD[i].Target < leftD[j].Target })
	sort.Slice(leftC, func(i, j int) bool { return leftC[i].Target < leftC[j].Target })
	n := min(len(leftD), len(leftC))
	for i := 0; i < n; i++ {
		plan.Updates = append(plan.Updates, Update{Desired: leftD[i], Current: leftC[i]})
	}
	plan.Creates = append(plan.Creates, leftD[n:]...)
	plan.Deletes = append(plan.Deletes, leftC[n:]...)
}
