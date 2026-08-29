package main

import (
	"reflect"
	"testing"
)

// TestResolveAllowedOrigins_CustomValue covers a valid CORS_ALLOWED_ORIGINS value
// being used verbatim, for a non-development env (the Staging use case this
// feature exists for).
func TestResolveAllowedOrigins_CustomValue(t *testing.T) {
	got, err := ResolveAllowedOrigins("https://mazadpay-validation-admin.onrender.com", "staging")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"https://mazadpay-validation-admin.onrender.com"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

// TestResolveAllowedOrigins_TrimAndDeduplicate covers whitespace trimming and
// duplicate removal across a comma-separated list.
func TestResolveAllowedOrigins_TrimAndDeduplicate(t *testing.T) {
	raw := "  https://a.example.com , https://b.example.com,https://a.example.com  ,https://b.example.com"
	got, err := ResolveAllowedOrigins(raw, "staging")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"https://a.example.com", "https://b.example.com"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v (expected trimmed + de-duplicated, first-occurrence order preserved)", got, want)
	}
}

// TestResolveAllowedOrigins_RejectsWildcard is the explicit safety requirement:
// "*" must never be accepted, especially since AllowCredentials-equivalent
// behavior (Authorization header) is enabled on this server's CORS config.
func TestResolveAllowedOrigins_RejectsWildcard(t *testing.T) {
	_, err := ResolveAllowedOrigins("*", "staging")
	if err == nil {
		t.Fatal("expected an error for wildcard origin \"*\", got nil")
	}

	_, err = ResolveAllowedOrigins("https://good.example.com,*", "staging")
	if err == nil {
		t.Fatal("expected an error when wildcard is mixed with a valid origin, got nil")
	}
}

// TestResolveAllowedOrigins_RejectsPathAndQuery covers rejection of an origin
// carrying a path, query string, or fragment -- only a bare scheme://host[:port]
// is a valid CORS origin.
func TestResolveAllowedOrigins_RejectsPathAndQuery(t *testing.T) {
	cases := []string{
		"https://example.com/admin",
		"https://example.com/",
		"https://example.com?x=1",
		"https://example.com#fragment",
		"https://user:pass@example.com",
	}
	for _, raw := range cases {
		t.Run(raw, func(t *testing.T) {
			_, err := ResolveAllowedOrigins(raw, "staging")
			if raw == "https://example.com/" {
				// A single trailing slash with no other path segment is
				// tolerated (browsers normalize Origin without a trailing
				// slash, but some manual entry may include one) -- still,
				// document the exact accepted behavior via this subtest
				// rather than assuming; verify it does NOT error.
				if err != nil {
					t.Fatalf("expected trailing-slash-only origin to be accepted, got error: %v", err)
				}
				return
			}
			if err == nil {
				t.Errorf("expected an error for invalid origin %q, got nil", raw)
			}
		})
	}
}

// TestResolveAllowedOrigins_RejectsNonHTTPScheme covers rejection of schemes other
// than http/https (e.g. ftp, javascript, data).
func TestResolveAllowedOrigins_RejectsNonHTTPScheme(t *testing.T) {
	cases := []string{"ftp://example.com", "javascript://example.com", "example.com"}
	for _, raw := range cases {
		t.Run(raw, func(t *testing.T) {
			_, err := ResolveAllowedOrigins(raw, "staging")
			if err == nil {
				t.Errorf("expected an error for non-http(s) origin %q, got nil", raw)
			}
		})
	}
}

// TestResolveAllowedOrigins_DevelopmentFallbackUnchanged verifies that when
// CORS_ALLOWED_ORIGINS is empty and APP_ENV=development, the exact previous
// localhost fallback is returned, unchanged.
func TestResolveAllowedOrigins_DevelopmentFallbackUnchanged(t *testing.T) {
	got, err := ResolveAllowedOrigins("", "development")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"http://localhost:5173", "http://localhost:3000"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

// TestResolveAllowedOrigins_ProductionFallbackUnchanged verifies that when
// CORS_ALLOWED_ORIGINS is empty and APP_ENV is anything other than "development"
// (including the existing production deployment, which has never set this
// variable and never will need to), the exact previous production allow-list is
// returned, unchanged.
func TestResolveAllowedOrigins_ProductionFallbackUnchanged(t *testing.T) {
	for _, env := range []string{"production", "", "prod", "unknown-value"} {
		t.Run(env, func(t *testing.T) {
			got, err := ResolveAllowedOrigins("", env)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			want := []string{
				"https://mazadpay-admin.onrender.com",
				"https://admin.mazadpay.com",
				"https://mazadpay.com",
				"https://www.mazadpay.com",
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("got %v, want %v", got, want)
			}
		})
	}
}

// TestResolveAllowedOrigins_StagingDoesNotTriggerDevelopmentFallback is the
// explicit regression test for the requirement that APP_ENV=staging must NOT
// fall back to the development (localhost) allow-list when CORS_ALLOWED_ORIGINS
// is empty -- it must fall back to the production list instead, same as any
// non-development env. Staging's actual origin must come from
// CORS_ALLOWED_ORIGINS being set explicitly (see
// TestResolveAllowedOrigins_CustomValue), never from an implicit env-name match.
func TestResolveAllowedOrigins_StagingDoesNotTriggerDevelopmentFallback(t *testing.T) {
	got, err := ResolveAllowedOrigins("", "staging")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, origin := range got {
		if origin == "http://localhost:5173" || origin == "http://localhost:3000" {
			t.Fatalf("APP_ENV=staging with empty CORS_ALLOWED_ORIGINS must not fall back to the development localhost list, got %v", got)
		}
	}
}

// TestResolveAllowedOrigins_EmptyAfterTrimIsError covers a value that is
// non-empty as a raw string but produces no usable origin after trimming/split
// (e.g. only commas/whitespace) -- this must fail loudly, not silently fall back.
func TestResolveAllowedOrigins_EmptyAfterTrimIsError(t *testing.T) {
	_, err := ResolveAllowedOrigins(" , , ", "staging")
	if err == nil {
		t.Fatal("expected an error for a CORS_ALLOWED_ORIGINS value with no usable origin after trimming, got nil")
	}
}

// TestValidateOrigin_DirectCases exercises validateOrigin directly for the
// documented scheme/host/path/query/fragment/userinfo rules.
func TestValidateOrigin_DirectCases(t *testing.T) {
	validCases := []string{
		"https://mazadpay-validation-admin.onrender.com",
		"http://localhost:5173",
		"https://example.com:8443",
	}
	for _, origin := range validCases {
		if err := validateOrigin(origin); err != nil {
			t.Errorf("validateOrigin(%q) unexpected error: %v", origin, err)
		}
	}

	invalidCases := []string{"*", "", "not a url", "https://", "ftp://x.com"}
	for _, origin := range invalidCases {
		if err := validateOrigin(origin); err == nil {
			t.Errorf("validateOrigin(%q) expected an error, got nil", origin)
		}
	}
}
