package main

import (
	"strings"
	"testing"
)

// TestBuildBackfillPlan_NoFullE164InOutput is the regression test for the PII leak
// fix: nothing derived from this package's public surface (planEntry.E164 as consumed
// by main.go's masked logging) should ever let a full E.164 number reach stdout. This
// test asserts on maskPhone() directly, since that's the only function allowed to
// touch an E.164 value before it's printed.
func TestMaskPhone_NeverReturnsFullNumber(t *testing.T) {
	cases := []string{"+22220123457", "+16135550123", "20123456", "1234"}
	for _, phone := range cases {
		masked := maskPhone(phone)
		if masked == phone {
			t.Fatalf("maskPhone(%q) returned the input unchanged — full number would leak", phone)
		}
		if strings.HasPrefix(masked, "+") {
			t.Fatalf("maskPhone(%q) = %q still starts with '+' — looks like an unmasked E.164 number", phone, masked)
		}
		if !strings.HasPrefix(masked, "####") {
			t.Fatalf("maskPhone(%q) = %q does not start with the expected #### mask", phone, masked)
		}
		// Only the last 4 digits of the ORIGINAL input should appear after the mask.
		wantSuffix := phone[len(phone)-4:]
		if !strings.HasSuffix(masked, wantSuffix) {
			t.Fatalf("maskPhone(%q) = %q does not end with the expected last-4 suffix %q", phone, masked, wantSuffix)
		}
	}
}

func mrCandidate(id, phone string) candidate { return candidate{ID: id, Phone: phone} }

// TestBuildBackfillPlan_DetectsConflictBetweenTwoFormats is the core new-requirement
// test: two different raw phone formats for the SAME logical Mauritanian number must
// both be flagged as conflicting, neither silently "wins".
func TestBuildBackfillPlan_DetectsConflictBetweenTwoFormats(t *testing.T) {
	candidates := []candidate{
		mrCandidate("user-a", "+22220123457"), // already E.164-prefixed
		mrCandidate("user-b", "20123457"),     // bare national number, same logical number
	}
	plan := buildBackfillPlan(candidates, existingE164Lookup{})

	if len(plan) != 2 {
		t.Fatalf("expected 2 plan entries, got %d", len(plan))
	}
	for _, p := range plan {
		if p.Status != "conflict" {
			t.Errorf("entry for %s: status = %q, want %q (both sides of a same-E164 collision must be marked conflict)", p.ID, p.Status, "conflict")
		}
	}
}

// TestBuildBackfillPlan_ExcludesAllConflictSides ensures no ordering/priority (e.g.
// created_at position in the input slice) picks a "winner" — every side is excluded.
func TestBuildBackfillPlan_ExcludesAllConflictSides(t *testing.T) {
	candidates := []candidate{
		mrCandidate("first-by-created-at", "20123458"),
		mrCandidate("second-by-created-at", "+22220123458"),
		mrCandidate("third-by-created-at", "20123458"), // duplicate raw value, still a 3-way collision
	}
	plan := buildBackfillPlan(candidates, existingE164Lookup{})

	migratable := 0
	conflict := 0
	for _, p := range plan {
		switch p.Status {
		case "migratable":
			migratable++
		case "conflict":
			conflict++
		}
	}
	if migratable != 0 {
		t.Errorf("expected 0 migratable entries among an all-conflicting group, got %d — some side was silently prioritized", migratable)
	}
	if conflict != len(candidates) {
		t.Errorf("expected all %d candidates marked conflict, got %d", len(candidates), conflict)
	}
}

// TestBuildBackfillPlan_ConflictsWithExistingRow covers collision against a row
// that was already normalized in a previous run (not just within the current batch).
func TestBuildBackfillPlan_ConflictsWithExistingRow(t *testing.T) {
	candidates := []candidate{
		mrCandidate("new-user", "20123459"),
	}
	existing := existingE164Lookup{
		"already-migrated-user": "+22220123459", // same logical number, already written
	}
	plan := buildBackfillPlan(candidates, existing)

	if len(plan) != 1 {
		t.Fatalf("expected 1 plan entry, got %d", len(plan))
	}
	if plan[0].Status != "conflict" {
		t.Errorf("status = %q, want %q — a candidate colliding with an already-normalized existing row must be a conflict, not silently migrated", plan[0].Status, "conflict")
	}
}

// TestBuildBackfillPlan_MigratableWhenNoConflict is the baseline positive case —
// distinct valid MR numbers with no collision must still be planned normally.
func TestBuildBackfillPlan_MigratableWhenNoConflict(t *testing.T) {
	candidates := []candidate{
		mrCandidate("user-x", "20123460"),
		mrCandidate("user-y", "30123461"),
	}
	plan := buildBackfillPlan(candidates, existingE164Lookup{})

	for _, p := range plan {
		if p.Status != "migratable" {
			t.Errorf("entry for %s: status = %q, want %q", p.ID, p.Status, "migratable")
		}
	}
}

// TestBuildBackfillPlan_InvalidNumberSkipped preserves existing skip behavior.
func TestBuildBackfillPlan_InvalidNumberSkipped(t *testing.T) {
	candidates := []candidate{mrCandidate("bad-number", "0000")}
	plan := buildBackfillPlan(candidates, existingE164Lookup{})

	if len(plan) != 1 || plan[0].Status != "skip_unparseable" {
		t.Fatalf("expected 1 skip_unparseable entry, got %+v", plan)
	}
}

// TestSummarize_CountsMatchStatuses is a direct test of the PII-free summary line
// fields requested (scanned/migratable/skipped_invalid/conflicts).
func TestSummarize_CountsMatchStatuses(t *testing.T) {
	plan := []planEntry{
		{Status: "migratable"},
		{Status: "migratable"},
		{Status: "skip_unparseable"},
		{Status: "skip_unexpected_region"},
		{Status: "conflict"},
		{Status: "conflict"},
		{Status: "conflict"},
	}
	s := summarize(plan)

	if s.Scanned != 7 {
		t.Errorf("Scanned = %d, want 7", s.Scanned)
	}
	if s.Migratable != 2 {
		t.Errorf("Migratable = %d, want 2", s.Migratable)
	}
	if s.SkippedInvalid != 2 {
		t.Errorf("SkippedInvalid = %d, want 2", s.SkippedInvalid)
	}
	if s.Conflicts != 3 {
		t.Errorf("Conflicts = %d, want 3", s.Conflicts)
	}
}

// TestBuildBackfillPlan_IdempotencyAcrossRuns simulates re-running the tool after a
// first successful (simulated) execute: candidates already written no longer appear
// in the input (WHERE phone_e164 IS NULL upstream), so a second call with only the
// remaining candidates plus the newly-existing row must not re-flag anything as a
// fresh conflict or duplicate work.
func TestBuildBackfillPlan_IdempotencyAcrossRuns(t *testing.T) {
	// First run: two independent, non-conflicting candidates.
	firstRun := []candidate{
		mrCandidate("user-p", "20123470"),
		mrCandidate("user-q", "30123471"),
	}
	plan1 := buildBackfillPlan(firstRun, existingE164Lookup{})
	for _, p := range plan1 {
		if p.Status != "migratable" {
			t.Fatalf("first run: entry for %s unexpectedly not migratable: %+v", p.ID, p)
		}
	}

	// Simulate: user-p was successfully written (WHERE phone_e164 IS NULL excludes it
	// from the next SELECT), user-q was NOT (e.g. it failed or the run was partial).
	existingAfterFirstRun := existingE164Lookup{"user-p": plan1[0].E164}
	secondRunCandidates := []candidate{mrCandidate("user-q", "30123471")}

	plan2 := buildBackfillPlan(secondRunCandidates, existingAfterFirstRun)
	if len(plan2) != 1 || plan2[0].Status != "migratable" {
		t.Fatalf("second run: expected user-q still migratable and untouched by user-p's prior result, got %+v", plan2)
	}
}
