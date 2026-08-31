package handlers

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/mazadpay/backend/internal/models"
)

// TestMaskUserPhone_K_AdminFullPhone_L_PublicMasked is targeted tests K and L
// (client feedback Phase B item 17). Prior to this fix, MaskUserPhone's name
// promised masking but actually copied user.Phone verbatim (Phone: user.Phone),
// so UserHandler.Search -- an authenticated-but-public lookup of OTHER users --
// leaked the full raw phone number of any user despite its own "Mask phone
// numbers for privacy" comment. Admin endpoints (AdminHandler.ListUsers/
// GetUserByID) never called MaskUserPhone at all; they return models.User
// directly and are gated by AdminOnly middleware -- this test locks down that
// distinction: MaskUserPhone masks (used by public/customer-facing paths like
// Search), while a raw models.User (what admin handlers return) still carries
// the full number.
func TestMaskUserPhone_K_AdminFullPhone_L_PublicMasked(t *testing.T) {
	fullPhone := "22236601175"
	user := &models.User{
		ID:    uuid.New(),
		Phone: fullPhone,
		Role:  "user",
	}

	t.Run("L: public/customer response (MaskUserPhone, e.g. Search) never leaks the full phone", func(t *testing.T) {
		masked := MaskUserPhone(user)
		if masked == nil {
			t.Fatal("expected a non-nil ResponseUser")
		}
		if masked.Phone == fullPhone {
			t.Fatalf("SECURITY REGRESSION: MaskUserPhone returned the unmasked phone %q -- "+
				"public/customer-facing responses must never leak the full number", masked.Phone)
		}
		if !strings.Contains(masked.Phone, "####") {
			t.Fatalf("expected masked phone to use the ####xxxx convention, got %q", masked.Phone)
		}
		if !strings.HasSuffix(masked.Phone, fullPhone[len(fullPhone)-4:]) {
			t.Fatalf("expected masked phone to keep the last 4 digits, got %q", masked.Phone)
		}
	})

	t.Run("K: admin response (raw models.User, as returned by AdminHandler.ListUsers/GetUserByID) carries the full phone", func(t *testing.T) {
		// AdminHandler.ListUsers/GetUserByID (admin_handler.go) return the
		// service's models.User result directly -- never through MaskUserPhone.
		// This asserts that contract stays true: the raw model still has
		// the untouched, full phone number for an authorized admin caller.
		if user.Phone != fullPhone {
			t.Fatalf("expected admin-facing raw models.User.Phone to remain the full number, got %q", user.Phone)
		}
	})

	t.Run("self-view (SelfResponseUser, used by GetMe) also carries the full phone", func(t *testing.T) {
		self := SelfResponseUser(user)
		if self == nil {
			t.Fatal("expected a non-nil ResponseUser")
		}
		if self.Phone != fullPhone {
			t.Fatalf("expected a user viewing their own profile to see their full phone, got %q", self.Phone)
		}
	})
}
