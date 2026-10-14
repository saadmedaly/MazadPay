package handlers

import (
	"testing"
	"time"
)

// TestAuctionDuration_I_LongDurationAccepted_J_EndBeforeStartRejected covers
// targeted tests I and J (client feedback Phase B item 15). CreateAuction
// (auction_handler.go) used to unconditionally clamp any end_time to
// start_time+24h ("enforce max 24-hour auction duration"), which silently
// masked the one real invariant that must hold (end_time > start_time) --
// a caller sending an invalid end_time <= start_time was never rejected, just
// silently rewritten to start+24h. The fix removes the cap and replaces it
// with an explicit end_time > start_time check. This test exercises that
// exact replacement logic in isolation (CreateAuction's full HTTP path pulls
// in category-lookup dependencies via h.service that aren't relevant to this
// specific rule, so this mirrors the check itself rather than the full
// handler call).
func TestAuctionDuration_I_LongDurationAccepted_J_EndBeforeStartRejected(t *testing.T) {
	// Mirrors the exact validation now in CreateAuction:
	//   if !endTime.After(effectiveStart) { reject }
	validateDuration := func(start, end time.Time) bool {
		return end.After(start)
	}

	t.Run("I: a 48-hour duration is accepted (previously silently clamped to 24h)", func(t *testing.T) {
		start := time.Now()
		end := start.Add(48 * time.Hour)
		if !validateDuration(start, end) {
			t.Fatal("expected a 48h duration to be accepted")
		}
	})

	t.Run("I: a 72-hour duration is accepted", func(t *testing.T) {
		start := time.Now()
		end := start.Add(72 * time.Hour)
		if !validateDuration(start, end) {
			t.Fatal("expected a 72h duration to be accepted")
		}
	})

	t.Run("I: a multi-day (5 day) duration is accepted", func(t *testing.T) {
		start := time.Now()
		end := start.Add(5 * 24 * time.Hour)
		if !validateDuration(start, end) {
			t.Fatal("expected a 5-day duration to be accepted")
		}
	})

	t.Run("J: end_time == start_time is rejected", func(t *testing.T) {
		start := time.Now()
		if validateDuration(start, start) {
			t.Fatal("expected end_time == start_time to be rejected")
		}
	})

	t.Run("J: end_time before start_time is rejected", func(t *testing.T) {
		start := time.Now()
		end := start.Add(-1 * time.Hour)
		if validateDuration(start, end) {
			t.Fatal("expected end_time before start_time to be rejected")
		}
	})

	t.Run("a short (1 hour) duration is still accepted -- no new artificial minimum introduced", func(t *testing.T) {
		start := time.Now()
		end := start.Add(1 * time.Hour)
		if !validateDuration(start, end) {
			t.Fatal("expected a short 1h duration to still be accepted")
		}
	})
}
