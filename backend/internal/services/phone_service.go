package services

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/nyaruka/phonenumbers"
)

// NormalizeE164 parses rawNumber and validates it against countryISO — an ISO-2 region
// code such as "MR", "TN", "US", "CA" — which is now REQUIRED and authoritative (Product
// Review / International Auth Phase 2 correction).
//
// A dial code alone (e.g. "+1") is NOT sufficient to identify a country: many countries
// share the same calling code (the North American Numbering Plan alone covers US, CA,
// and a dozen Caribbean nations under +1). Before this fix, a number starting with "+"
// bypassed any region hint entirely and let libphonenumber's own best-guess region
// detection decide — which for a shared dial code can silently pick the wrong country
// (e.g. a Canadian number silently stored as US). This function now ALWAYS validates the
// parsed number specifically against countryISO via phonenumbers.IsValidNumberForRegion,
// and rejects the request if the number is not a valid number for that specific region —
// even if it happens to be a validly-formatted number for some other country sharing the
// same calling code. The caller-supplied countryISO becomes the source of truth for which
// country a phone number belongs to; libphonenumber is only used to validate the number
// against that specific country's numbering plan, never to silently override it.
func NormalizeE164(rawNumber, countryISO string) (e164 string, isoRegion string, err error) {
	rawNumber = strings.TrimSpace(rawNumber)
	if rawNumber == "" {
		return "", "", fmt.Errorf("empty phone number")
	}

	region := resolveRegionHint(countryISO)
	if region == "" {
		return "", "", fmt.Errorf("country_iso is required and must be a recognized ISO-2 region code (got %q)", countryISO)
	}

	num, parseErr := phonenumbers.Parse(rawNumber, region)
	if parseErr != nil {
		return "", "", fmt.Errorf("failed to parse phone number: %w", parseErr)
	}

	// Authoritative check: the number must be valid specifically FOR the caller-declared
	// region, not merely valid for whatever region the raw digits happen to resolve to.
	// This is what prevents a US number the client tagged as "CA" (or vice versa, or any
	// other shared-dial-code mix-up) from being silently accepted under the wrong country.
	if !phonenumbers.IsValidNumberForRegion(num, region) {
		return "", "", fmt.Errorf("phone number is not valid for region %s", region)
	}

	e164 = phonenumbers.Format(num, phonenumbers.E164)
	isoRegion = region

	return e164, isoRegion, nil
}

// resolveRegionHint turns countryISO into an ISO-2 region code phonenumbers can use.
// Accepts either an ISO-2 code directly (e.g. "MR", "mr") or — for backward
// compatibility with any caller still passing a dial code — a dial code such as "+1",
// resolved to phonenumbers' own "main" region for that calling code. Note that resolving
// a dial code this way is inherently ambiguous for shared codes (e.g. "+1" always
// resolves to "US", never "CA") — callers SHOULD pass an explicit ISO-2 region whenever
// possible; the dial-code fallback exists only so an old/malformed caller doesn't hard
// crash, not as a recommended path. Returns "" if it cannot be resolved.
func resolveRegionHint(countryISO string) string {
	countryISO = strings.TrimSpace(countryISO)
	if countryISO == "" {
		return ""
	}

	// Already an ISO-2 region code (e.g. "MR", "mr"). GetCountryCodeForRegion returns 0
	// for a region phonenumbers doesn't recognize, which we treat as "not a valid hint".
	if len(countryISO) == 2 {
		upper := strings.ToUpper(countryISO)
		if phonenumbers.GetCountryCodeForRegion(upper) != 0 {
			return upper
		}
	}

	// Dial code form, e.g. "+222" or "222" — ambiguous for shared codes, kept only for
	// defensiveness against callers that haven't been updated to send country_iso.
	digits := strings.TrimPrefix(countryISO, "+")
	callingCode, convErr := strconv.Atoi(digits)
	if convErr != nil {
		return ""
	}

	return phonenumbers.GetRegionCodeForCountryCode(callingCode)
}
