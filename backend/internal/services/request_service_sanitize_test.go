package services

import (
	"strings"
	"testing"

	"github.com/mazadpay/backend/internal/models"
)

// TestSanitizeDescriptions_XSS is the regression test for the security-review
// requirement that description sanitization actually neutralizes malicious HTML/script
// content rather than merely stripping tags in a way that could still be exploited (e.g.
// by an admin dashboard rendering a "cleaned" description as raw HTML).
// bluemonday.StrictPolicy() strips ALL HTML tags — the sanitized output must contain no
// "<" character at all, and specifically must not contain any of the classic XSS vectors.
func TestSanitizeDescriptions_XSS(t *testing.T) {
	cases := []struct {
		name  string
		input string
	}{
		{"script tag", `<script>alert('xss')</script>hello`},
		{"img onerror", `<img src=x onerror=alert(1)>`},
		{"svg onload", `<svg onload=alert(1)>`},
		{"javascript: href", `<a href="javascript:alert(1)">click</a>`},
		{"nested/broken tags", `<scr<script>ipt>alert(1)</scr</script>ipt>`},
		{"event handler on plain text wrapper", `<div onclick="alert(1)">text</div>`},
		{"iframe injection", `<iframe src="javascript:alert(1)"></iframe>`},
		{"style expression", `<style>body{background:url("javascript:alert(1)")}</style>`},
		{"encoded script-like text stays as harmless text", `&lt;script&gt;alert(1)&lt;/script&gt;`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := &models.AuctionRequest{
				DescriptionAr: stringPtr(tc.input),
				DescriptionFr: stringPtr(tc.input),
				DescriptionEn: stringPtr(tc.input),
			}
			sanitizeDescriptions(req)

			for label, got := range map[string]*string{
				"ar": req.DescriptionAr,
				"fr": req.DescriptionFr,
				"en": req.DescriptionEn,
			} {
				if got == nil {
					t.Fatalf("[%s] description became nil after sanitization", label)
				}
				lower := strings.ToLower(*got)
				if strings.Contains(*got, "<") || strings.Contains(*got, ">") {
					t.Errorf("[%s] sanitized output still contains angle brackets: %q", label, *got)
				}
				if strings.Contains(lower, "onerror") || strings.Contains(lower, "onload") || strings.Contains(lower, "onclick") {
					t.Errorf("[%s] sanitized output still contains an event handler attribute: %q", label, *got)
				}
				if strings.Contains(lower, "javascript:") {
					t.Errorf("[%s] sanitized output still contains a javascript: URI: %q", label, *got)
				}
				if strings.Contains(lower, "<script") {
					t.Errorf("[%s] sanitized output still contains a script tag: %q", label, *got)
				}
			}
		})
	}
}

// TestSanitizeDescriptions_PreservesPlainText ensures the sanitizer doesn't mangle
// legitimate plain-text descriptions (Arabic, French, English, punctuation, newlines) —
// a sanitizer that's "safe" but destroys normal input isn't actually usable.
func TestSanitizeDescriptions_PreservesPlainText(t *testing.T) {
	plainAr := "دلة وعاء زجاجي أثرية، صناعة أوروبية عريقة، حالة ممتازة."
	plainFr := "Théière antique en verre, fabrication européenne, excellent état."
	plainEn := "Antique glass teapot, fine European craftsmanship, excellent condition."

	req := &models.AuctionRequest{
		DescriptionAr: stringPtr(plainAr),
		DescriptionFr: stringPtr(plainFr),
		DescriptionEn: stringPtr(plainEn),
	}
	sanitizeDescriptions(req)

	if req.DescriptionAr == nil || *req.DescriptionAr != plainAr {
		t.Errorf("Arabic plain text was altered: got %q, want %q", derefOrNil(req.DescriptionAr), plainAr)
	}
	if req.DescriptionFr == nil || *req.DescriptionFr != plainFr {
		t.Errorf("French plain text was altered: got %q, want %q", derefOrNil(req.DescriptionFr), plainFr)
	}
	if req.DescriptionEn == nil || *req.DescriptionEn != plainEn {
		t.Errorf("English plain text was altered: got %q, want %q", derefOrNil(req.DescriptionEn), plainEn)
	}
}

func derefOrNil(s *string) string {
	if s == nil {
		return "<nil>"
	}
	return *s
}
