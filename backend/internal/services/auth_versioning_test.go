package services

import "testing"

// TestValidCountryCodesLegacy_ExactlyBaselineFour locks the v1 (/v1/api/auth/...)
// allow-list to exactly the 4 dial codes accepted by the version currently published on
// Google Play. This is a regression test for the API Versioning fix: v1 must NEVER
// silently gain a new country, since that would mean v1's behavior drifted from what the
// published app was built against — any new country belongs exclusively in v2
// (Register/NormalizeE164), never in this map.
func TestValidCountryCodesLegacy_ExactlyBaselineFour(t *testing.T) {
	want := map[string]string{
		"+222": "MR",
		"+221": "SN",
		"+212": "MA",
		"+216": "TN",
	}

	if len(ValidCountryCodesLegacy) != len(want) {
		t.Fatalf("ValidCountryCodesLegacy has %d entries, want exactly %d", len(ValidCountryCodesLegacy), len(want))
	}
	for code, iso := range want {
		got, ok := ValidCountryCodesLegacy[code]
		if !ok {
			t.Errorf("ValidCountryCodesLegacy is missing dial code %q (present in the published baseline)", code)
			continue
		}
		if got != iso {
			t.Errorf("ValidCountryCodesLegacy[%q] = %q, want %q", code, got, iso)
		}
	}

	// A country added in v2 (e.g. Saudi Arabia, +966) must NOT leak into v1.
	for _, unexpected := range []string{"+966", "+1", "+33", "+20"} {
		if _, present := ValidCountryCodesLegacy[unexpected]; present {
			t.Errorf("ValidCountryCodesLegacy unexpectedly contains %q — v1 must stay frozen to the 4-country baseline, new countries belong in v2 only", unexpected)
		}
	}
}
