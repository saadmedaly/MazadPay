package main

import (
	"fmt"
	"net/url"
	"strings"
)

// productionAllowedOrigins is the exact, unchanged production allow-list — moved
// here verbatim from the previous hardcoded value in main(), so behavior for
// existing deployments (APP_ENV anything other than "development", with
// CORS_ALLOWED_ORIGINS unset) is byte-for-byte identical to before this change.
var productionAllowedOrigins = []string{
	"https://mazadpay-admin.onrender.com",
	"https://admin.mazadpay.com",
	"https://mazadpay.com",
	"https://www.mazadpay.com",
}

// developmentAllowedOrigins is the exact, unchanged development allow-list.
var developmentAllowedOrigins = []string{
	"http://localhost:5173",
	"http://localhost:3000",
}

// ResolveAllowedOrigins determines the CORS allow-list for the server, given the
// raw (unparsed) CORS_ALLOWED_ORIGINS environment value and the current APP_ENV.
//
// Precedence (CORS_ALLOWED_ORIGINS Phase 1):
//  1. rawOrigins non-empty (after trim) -> parse and validate it, use it verbatim
//     for ANY environment (including "development" and a new "staging" value) —
//     this is what lets a Staging deployment (e.g. mazadpay-validation-api) allow
//     its own Admin Staging origin without touching the hardcoded production list.
//  2. rawOrigins empty AND env == "development" -> the exact previous
//     development fallback (localhost), unchanged.
//  3. rawOrigins empty AND any other env (including a new "staging" env value
//     that doesn't set CORS_ALLOWED_ORIGINS) -> the exact previous production
//     fallback, unchanged. This is what guarantees production's default behavior
//     never changes: production has never set CORS_ALLOWED_ORIGINS and never
//     will need to.
//
// Returns an error (never partial success) if rawOrigins is non-empty but contains
// any invalid entry — see validateOrigin. The caller (main) must treat this as a
// fatal startup error, never fall back to a wide-open or empty CORS policy.
func ResolveAllowedOrigins(rawOrigins, env string) ([]string, error) {
	trimmed := strings.TrimSpace(rawOrigins)
	if trimmed == "" {
		if env == "development" {
			return developmentAllowedOrigins, nil
		}
		return productionAllowedOrigins, nil
	}

	parts := strings.Split(trimmed, ",")
	seen := make(map[string]bool, len(parts))
	origins := make([]string, 0, len(parts))

	for _, raw := range parts {
		origin := strings.TrimSpace(raw)
		if origin == "" {
			// A stray trailing/leading comma or double comma produces an empty
			// segment after trim -- skip it rather than treat it as an error,
			// this is a common, harmless formatting slip.
			continue
		}
		if err := validateOrigin(origin); err != nil {
			return nil, fmt.Errorf("CORS_ALLOWED_ORIGINS: invalid origin %q: %w", origin, err)
		}
		if seen[origin] {
			continue // de-duplicate, keep first occurrence's position
		}
		seen[origin] = true
		origins = append(origins, origin)
	}

	if len(origins) == 0 {
		return nil, fmt.Errorf("CORS_ALLOWED_ORIGINS is set but contains no usable origin after trimming")
	}

	return origins, nil
}

// validateOrigin enforces that a single CORS origin entry is a bare
// "scheme://host[:port]" with no path, query string, or fragment, using only the
// http or https scheme, and is never the wildcard "*" -- this server always sends
// AllowCredentials-equivalent behavior (Authorization header is in AllowHeaders),
// and a wildcard origin combined with credentialed requests is both rejected by
// browsers and a real security misconfiguration if it were ever accepted.
func validateOrigin(origin string) error {
	if origin == "*" {
		return fmt.Errorf("wildcard origin \"*\" is not allowed")
	}

	u, err := url.Parse(origin)
	if err != nil {
		return fmt.Errorf("not a valid URL: %w", err)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("scheme must be http or https, got %q", u.Scheme)
	}
	if u.Host == "" {
		return fmt.Errorf("missing host")
	}
	if u.Path != "" && u.Path != "/" {
		return fmt.Errorf("must be an origin only, no path")
	}
	if u.RawQuery != "" {
		return fmt.Errorf("must be an origin only, no query string")
	}
	if u.Fragment != "" {
		return fmt.Errorf("must be an origin only, no fragment")
	}
	if u.User != nil {
		return fmt.Errorf("must be an origin only, no userinfo")
	}

	return nil
}
