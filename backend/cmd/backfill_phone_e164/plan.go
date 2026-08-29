package main

import "github.com/mazadpay/backend/internal/services"

// candidate is a users row that still needs phone_e164/phone_country_iso backfilled.
type candidate struct {
	ID    string
	Phone string
}

// planEntry is the outcome for one candidate after normalization: either migratable,
// skipped (unparseable/wrong region), or conflicting with another candidate/existing
// row (never both skipped and migratable).
type planEntry struct {
	ID     string
	E164   string
	ISO    string
	Status string // "migratable" | "skip_unparseable" | "skip_unexpected_region" | "conflict"
	Err    error  // set for skip_unparseable
}

// existingE164Lookup reports whether e164 is already present on a DIFFERENT user row
// already in the database (phone_e164 IS NOT NULL). The caller supplies this as a
// plain map (id -> e164) so the conflict-detection logic below stays pure/testable
// without a DB dependency.
type existingE164Lookup map[string]string // user id -> phone_e164, for rows where phone_e164 IS NOT NULL already

// buildBackfillPlan normalizes every candidate via NormalizeE164(phone, "MR") and
// classifies each one, WITHOUT deciding any update order or picking a "winner" among
// conflicting rows — every side of a phone_e164 collision (whether the collision is
// between two candidates in this same run, or between a candidate and an already-
// normalized existing row) is marked "conflict" and excluded from the update plan.
// This function performs no I/O and is fully unit-testable.
func buildBackfillPlan(candidates []candidate, existing existingE164Lookup) []planEntry {
	// First pass: normalize everything, without yet knowing about collisions.
	type normalized struct {
		id, phone, e164, iso string
		err                  error
		unexpectedRegion     bool
	}
	norm := make([]normalized, 0, len(candidates))
	for _, c := range candidates {
		e164, iso, err := services.NormalizeE164(c.Phone, "MR")
		if err != nil {
			norm = append(norm, normalized{id: c.ID, phone: c.Phone, err: err})
			continue
		}
		if iso != "MR" {
			norm = append(norm, normalized{id: c.ID, phone: c.Phone, e164: e164, iso: iso, unexpectedRegion: true})
			continue
		}
		norm = append(norm, normalized{id: c.ID, phone: c.Phone, e164: e164, iso: iso})
	}

	// Second pass: detect collisions among the successfully-normalized candidates
	// themselves (two different raw phone values converging on the same E.164), and
	// against already-normalized rows already in the DB (existing).
	countByE164 := make(map[string]int)
	for _, n := range norm {
		if n.err != nil || n.unexpectedRegion {
			continue
		}
		countByE164[n.e164]++
	}
	// An existing row with the same e164 as a candidate is ALSO a conflict for that
	// candidate — even if this is the only candidate producing that e164 among the
	// current batch, writing it would violate idx_users_phone_e164_unique.
	existingByE164 := make(map[string]bool, len(existing))
	for _, e164 := range existing {
		existingByE164[e164] = true
	}

	plan := make([]planEntry, 0, len(candidates))
	for _, n := range norm {
		switch {
		case n.err != nil:
			plan = append(plan, planEntry{ID: n.id, Status: "skip_unparseable", Err: n.err})
		case n.unexpectedRegion:
			plan = append(plan, planEntry{ID: n.id, E164: n.e164, ISO: n.iso, Status: "skip_unexpected_region"})
		case countByE164[n.e164] > 1 || existingByE164[n.e164]:
			// Every side of the collision is marked conflict — no ordering/priority
			// (e.g. created_at, role) is applied here. A human resolves conflicts
			// separately.
			plan = append(plan, planEntry{ID: n.id, E164: n.e164, ISO: n.iso, Status: "conflict"})
		default:
			plan = append(plan, planEntry{ID: n.id, E164: n.e164, ISO: n.iso, Status: "migratable"})
		}
	}
	return plan
}

// summary is the PII-free aggregate counts requested for the final report line.
type summary struct {
	Scanned        int
	Migratable     int
	SkippedInvalid int
	Conflicts      int
	Updated        int
	Failed         int
}

func summarize(plan []planEntry) summary {
	s := summary{Scanned: len(plan)}
	for _, p := range plan {
		switch p.Status {
		case "migratable":
			s.Migratable++
		case "skip_unparseable", "skip_unexpected_region":
			s.SkippedInvalid++
		case "conflict":
			s.Conflicts++
		}
	}
	return s
}

// maskPhone masks a raw or E.164 phone number for logs: "####" + last 4 digits only,
// consistent with models.User.MaskPhone(). Used for EVERY phone value printed by this
// tool — raw phone, E.164, in both dry-run and execute, including error paths — never
// print an unmasked number under any code path.
func maskPhone(phone string) string {
	if len(phone) < 4 {
		return "####"
	}
	return "####" + phone[len(phone)-4:]
}

// shortID returns the first 8 characters of a UUID, never the full identifier.
func shortID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}
