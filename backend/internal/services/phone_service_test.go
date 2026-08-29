package services

import "testing"

func TestNormalizeE164_ValidNumbers(t *testing.T) {
	cases := []struct {
		name       string
		rawNumber  string
		countryISO string
		wantE164   string
		wantISO    string
	}{
		{"Mauritania with plus prefix", "+22220123456", "MR", "+22220123456", "MR"},
		{"Mauritania bare national number", "20123456", "MR", "+22220123456", "MR"},
		{"Tunisia with plus prefix", "+21620123456", "TN", "+21620123456", "TN"},
		{"Tunisia bare national number", "20123456", "TN", "+21620123456", "TN"},
		{"Senegal with plus prefix", "+221701234567", "SN", "+221701234567", "SN"},
		{"Morocco with plus prefix", "+212612345678", "MA", "+212612345678", "MA"},
		{"France with plus prefix", "+33612345678", "FR", "+33612345678", "FR"},
		{"United States with plus prefix", "+14155552671", "US", "+14155552671", "US"},
		{"Saudi Arabia with plus prefix", "+966512345678", "SA", "+966512345678", "SA"},
		// Canada shares the +1 dial code with the US and a dozen Caribbean nations —
		// a genuinely Canadian number must still normalize to region CA, not US, when
		// the caller declares CA explicitly.
		{"Canada with plus prefix", "+16135550123", "CA", "+16135550123", "CA"},
		{"Canada bare national number", "6135550123", "CA", "+16135550123", "CA"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e164, iso, err := NormalizeE164(tc.rawNumber, tc.countryISO)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if e164 != tc.wantE164 {
				t.Errorf("e164 = %q, want %q", e164, tc.wantE164)
			}
			if iso != tc.wantISO {
				t.Errorf("iso = %q, want %q", iso, tc.wantISO)
			}
		})
	}
}

// TestNormalizeE164_SharedDialCodeCountryMismatch is the regression test for the
// international-auth Phase 2 fix: a dial code alone (e.g. "+1") is not sufficient to
// identify a country, since the US, Canada, and many Caribbean nations all share it.
// Before the fix, NormalizeE164 ignored the region hint entirely whenever the raw
// number started with "+", and let libphonenumber's own best-guess region detection
// decide — which can silently store a Canadian number as "US" or vice versa. The fixed
// behavior must validate the number specifically against the caller-declared region and
// reject it (not silently accept it under a different region) when it doesn't match.
func TestNormalizeE164_SharedDialCodeCountryMismatch(t *testing.T) {
	cases := []struct {
		name       string
		rawNumber  string
		countryISO string
	}{
		// A real Toronto (Canada) number, declared as US: NANP-valid numbers are
		// frequently valid-shaped for *some* +1 region, but IsValidNumberForRegion must
		// reject it specifically against "US" since the number itself is CA-assigned.
		{"Canadian number declared as US", "+16135550123", "US"},
		// A number with a US-only area code (212 = New York), declared as CA.
		{"US number declared as CA", "+12125550123", "CA"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, iso, err := NormalizeE164(tc.rawNumber, tc.countryISO)
			if err == nil {
				t.Fatalf("expected an error for %q declared as region %q, but it was accepted (resolved region: %q) — this is the exact silent-country-mixup bug this test guards against", tc.rawNumber, tc.countryISO, iso)
			}
		})
	}
}

func TestNormalizeE164_InvalidNumbers(t *testing.T) {
	cases := []struct {
		name       string
		rawNumber  string
		countryISO string
	}{
		{"empty string", "", "MR"},
		{"too short for Mauritania", "+2221234", "MR"},
		{"garbage input", "not-a-phone-number", "MR"},
		{"missing country_iso entirely", "+22220123456", ""},
		{"unrecognized country_iso", "20123456", "ZZ"},
		{"country_iso that is not a real ISO-2 code", "20123456", "XX"},
		{"digits only, clearly invalid length for any region", "+1234567890123456789", "US"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := NormalizeE164(tc.rawNumber, tc.countryISO)
			if err == nil {
				t.Errorf("expected an error for input %q (country_iso %q), got none", tc.rawNumber, tc.countryISO)
			}
		})
	}
}

func TestNormalizeE164_DifferentFormatsSameNumber(t *testing.T) {
	// Numéro moritanien saisi de plusieurs façons différentes doit toujours
	// converger vers le même E.164 canonique — c'est la propriété qui empêche les
	// doublons de compte lors de l'inscription/connexion.
	inputs := []struct {
		raw    string
		region string
	}{
		{"+22220123456", "MR"},
		{"20123456", "MR"},
	}

	var want string
	for i, in := range inputs {
		got, _, err := NormalizeE164(in.raw, in.region)
		if err != nil {
			t.Fatalf("unexpected error normalizing %q/%q: %v", in.raw, in.region, err)
		}
		if i == 0 {
			want = got
			continue
		}
		if got != want {
			t.Errorf("normalization mismatch: input %q/%q produced %q, want %q", in.raw, in.region, got, want)
		}
	}
}
