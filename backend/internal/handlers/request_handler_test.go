package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/services"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// TestCreateAuctionRequest_ValidatorAcceptsDecimalPrices is the regression test for
// the Staging finding "Bad field type decimal.Decimal": go-playground/validator's
// built-in gt/gte tags on models.AuctionRequest's decimal.Decimal price fields
// (StartPrice, MinIncrement, InsuranceAmount) failed via reflection for any struct
// type it doesn't natively understand, which validator.Struct surfaced as an error
// causing the handler to return HTTP 500 for every create-auction-request call. The
// fix (newDecimalAwareValidator, request_validator.go) registers a CustomTypeFunc so
// validator can read a comparable value out of decimal.Decimal. This test proves
// validation no longer errors on a well-formed request carrying valid decimal prices.
func TestCreateAuctionRequest_ValidatorAcceptsDecimalPrices(t *testing.T) {
	h := NewRequestHandler(nil, nil)

	desc := "وصف تجريبي آمن يزيد عن عشرة أحرف بدون سكريبت"
	req := models.AuctionRequest{
		ID:              uuid.New(),
		UserID:          uuid.New(),
		CategoryID:      4,
		TitleAr:         "اختبار Staging",
		DescriptionAr:   &desc,
		StartPrice:      decimal.NewFromInt(100),
		MinIncrement:    decimal.NewFromInt(10),
		InsuranceAmount: decimal.Zero,
		StartDate:       time.Now(),
		EndDate:         time.Now().Add(24 * time.Hour),
		Status:          "pending",
	}

	if err := h.validate.Struct(req); err != nil {
		t.Fatalf("expected valid decimal prices to pass validation, got error: %v", err)
	}
}

// TestCreateAuctionRequest_ValidatorRejectsInvalidPrices proves the fix does not
// weaken validation: gt=0 (StartPrice, MinIncrement) and gte=0 (InsuranceAmount)
// must still reject out-of-range decimal values, not just avoid panicking/erroring.
func TestCreateAuctionRequest_ValidatorRejectsInvalidPrices(t *testing.T) {
	h := NewRequestHandler(nil, nil)
	desc := "وصف تجريبي آمن يزيد عن عشرة أحرف بدون سكريبت"

	baseReq := func() models.AuctionRequest {
		return models.AuctionRequest{
			ID:              uuid.New(),
			UserID:          uuid.New(),
			CategoryID:      4,
			TitleAr:         "اختبار Staging",
			DescriptionAr:   &desc,
			StartPrice:      decimal.NewFromInt(100),
			MinIncrement:    decimal.NewFromInt(10),
			InsuranceAmount: decimal.Zero,
			StartDate:       time.Now(),
			EndDate:         time.Now().Add(24 * time.Hour),
			Status:          "pending",
		}
	}

	cases := []struct {
		name   string
		mutate func(*models.AuctionRequest)
	}{
		{"zero start price", func(r *models.AuctionRequest) { r.StartPrice = decimal.Zero }},
		{"negative start price", func(r *models.AuctionRequest) { r.StartPrice = decimal.NewFromInt(-1) }},
		{"zero min increment", func(r *models.AuctionRequest) { r.MinIncrement = decimal.Zero }},
		{"negative min increment", func(r *models.AuctionRequest) { r.MinIncrement = decimal.NewFromInt(-5) }},
		{"negative insurance amount", func(r *models.AuctionRequest) { r.InsuranceAmount = decimal.NewFromInt(-1) }},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := baseReq()
			tc.mutate(&req)
			if err := h.validate.Struct(req); err == nil {
				t.Fatalf("expected validation to reject %s, got no error", tc.name)
			}
		})
	}
}

// TestDecimalAwareValidator_PrecisionEdgeCases proves newDecimalAwareValidator's use
// of decimal.Decimal.Sign() (an exact, big.Int-derived sign check) does not lose
// precision the way a float64 view (e.g. InexactFloat64()) could for values far
// outside float64's safe range — including an arbitrarily small non-zero decimal
// that a float conversion could round down to exactly 0.0.
func TestDecimalAwareValidator_PrecisionEdgeCases(t *testing.T) {
	v := newDecimalAwareValidator()

	type priceOnly struct {
		StartPrice      decimal.Decimal `validate:"gt=0"`
		InsuranceAmount decimal.Decimal `validate:"gte=0"`
	}

	tinyPositive, err := decimal.NewFromString("0.0000000000000000000000001")
	if err != nil {
		t.Fatalf("failed to construct tiny positive decimal: %v", err)
	}
	tinyNegative, err := decimal.NewFromString("-0.0000000000000000000000001")
	if err != nil {
		t.Fatalf("failed to construct tiny negative decimal: %v", err)
	}
	huge, err := decimal.NewFromString("99999999999999999999999999999999999999.99")
	if err != nil {
		t.Fatalf("failed to construct huge decimal: %v", err)
	}

	cases := []struct {
		name      string
		value     decimal.Decimal
		field     string // "start" validates gt=0, "insurance" validates gte=0
		wantValid bool
	}{
		{"tiny positive accepted for gt=0", tinyPositive, "start", true},
		{"tiny negative rejected for gt=0", tinyNegative, "start", false},
		{"zero rejected for gt=0", decimal.Zero, "start", false},
		{"zero accepted for gte=0", decimal.Zero, "insurance", true},
		{"huge positive accepted for gt=0", huge, "start", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := priceOnly{StartPrice: decimal.NewFromInt(1), InsuranceAmount: decimal.Zero}
			switch tc.field {
			case "start":
				p.StartPrice = tc.value
			case "insurance":
				p.InsuranceAmount = tc.value
			}

			err := v.Struct(p)
			if tc.wantValid && err != nil {
				t.Fatalf("expected %s (%s) to be valid, got error: %v", tc.name, tc.value.String(), err)
			}
			if !tc.wantValid && err == nil {
				t.Fatalf("expected %s (%s) to be rejected, got no error", tc.name, tc.value.String())
			}
		})
	}
}

// fakeRequestService is a minimal services.RequestService fake used only to prove
// the HTTP handler path (parsing -> binding -> validation -> service call) no longer
// 500s on decimal.Decimal fields. It does not touch a real database.
type fakeRequestService struct {
	services.RequestService
	createCalled bool
	createErr    error
}

func (f *fakeRequestService) CreateAuctionRequest(ctx context.Context, req *models.AuctionRequest) error {
	f.createCalled = true
	return f.createErr
}

// TestCreateAuctionRequest_HTTP_NoLongerReturns500OnDecimalFields is the HTTP-level
// regression test for the exact failure observed in Staging: POST
// /v1/api/requests/auctions with a well-formed body (valid category, description,
// decimal prices) returned HTTP 500 with "Bad field type decimal.Decimal" because
// validator.Struct errored before the request ever reached the service/repository
// layer. This test drives the real handler through Fiber's request/response cycle
// (body parsing, JWT-context simulation, h.validate.Struct, then the service call)
// using a fake RequestService in place of a database, and asserts the response is no
// longer a 500 caused by decimal validation.
func TestCreateAuctionRequest_HTTP_NoLongerReturns500OnDecimalFields(t *testing.T) {
	fakeSvc := &fakeRequestService{}
	h := NewRequestHandler(fakeSvc, zap.NewNop())

	app := fiber.New()
	userID := uuid.New()
	app.Post("/v1/api/requests/auctions", func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		return h.CreateAuctionRequest(c)
	})

	// user_id is included in the body because h.validate.Struct(req) (which carries
	// the "required" tag on models.AuctionRequest.UserID) runs before the handler
	// overwrites req.UserID from the authenticated context -- pre-existing handler
	// behavior, unrelated to and unchanged by this decimal-validation fix.
	body := map[string]interface{}{
		"user_id":          userID.String(),
		"category_id":      4,
		"title_ar":         "اختبار Staging",
		"description_ar":   "وصف تجريبي آمن يزيد عن عشرة أحرف بدون سكريبت",
		"start_price":      100,
		"min_increment":    10,
		"insurance_amount": 0,
		"start_date":       time.Now().Format(time.RFC3339),
		"end_date":         time.Now().Add(24 * time.Hour).Format(time.RFC3339),
	}
	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal request body: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/api/requests/auctions", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusInternalServerError {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected no 500 (the decimal.Decimal validation bug), got HTTP %d, body: %s", resp.StatusCode, body)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected HTTP 200 for a well-formed request, got HTTP %d, body: %s", resp.StatusCode, body)
	}
	if !fakeSvc.createCalled {
		t.Fatal("expected the request to reach the service layer (CreateAuctionRequest), but it did not")
	}
}

// TestCreateAuctionRequest_HTTP_StillRejectsInvalidPrice proves the HTTP path still
// enforces gt=0 on start_price after the fix -- the decimal-aware validator must not
// let an invalid price through to the service layer.
func TestCreateAuctionRequest_HTTP_StillRejectsInvalidPrice(t *testing.T) {
	fakeSvc := &fakeRequestService{}
	h := NewRequestHandler(fakeSvc, zap.NewNop())

	app := fiber.New()
	userID := uuid.New()
	app.Post("/v1/api/requests/auctions", func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		return h.CreateAuctionRequest(c)
	})

	body := map[string]interface{}{
		"user_id":          userID.String(),
		"category_id":      4,
		"title_ar":         "اختبار Staging",
		"description_ar":   "وصف تجريبي آمن يزيد عن عشرة أحرف بدون سكريبت",
		"start_price":      0,
		"min_increment":    10,
		"insurance_amount": 0,
		"start_date":       time.Now().Format(time.RFC3339),
		"end_date":         time.Now().Add(24 * time.Hour).Format(time.RFC3339),
	}
	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal request body: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/api/requests/auctions", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected HTTP 400 for start_price=0 (gt=0 violation), got HTTP %d", resp.StatusCode)
	}
	if fakeSvc.createCalled {
		t.Fatal("expected an invalid price to be rejected before reaching the service layer, but the service was called")
	}
}
