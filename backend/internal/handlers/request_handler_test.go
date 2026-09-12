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
// 500s on decimal.Decimal fields, and to observe exactly which UserID reaches the
// service layer (for the user-id spoofing regression test). It does not touch a
// real database.
type fakeRequestService struct {
	services.RequestService
	createCalled   bool
	createErr      error
	receivedUserID uuid.UUID

	createBannerCalled   bool
	createBannerErr      error
	receivedBannerUserID uuid.UUID
	receivedBannerReq    *models.BannerRequest

	auctionRequestsResult []models.AuctionRequest
	auctionRequestsTotal  int
	auctionRequestsErr    error

	bannerRequestsResult []models.BannerRequest
	bannerRequestsTotal  int
	bannerRequestsErr    error

	updateCalled         bool
	updateErr            error
	updateReceivedUserID uuid.UUID

	adminUpdateCalled bool
	adminUpdateErr    error
}

func (f *fakeRequestService) CreateAuctionRequest(ctx context.Context, req *models.AuctionRequest) error {
	f.createCalled = true
	f.receivedUserID = req.UserID
	return f.createErr
}

// CreateBannerRequest lets a test observe exactly which UserID and which
// fields (target_url, starts_at, ends_at) reach the service layer -- for
// Bug A's regression tests (client feedback: banner/ad request submission
// returning "Invalid request body").
func (f *fakeRequestService) CreateBannerRequest(ctx context.Context, req *models.BannerRequest) error {
	f.createBannerCalled = true
	f.receivedBannerUserID = req.UserID
	f.receivedBannerReq = req
	return f.createBannerErr
}

// UpdateAuctionRequest/AdminUpdateAuctionRequest let a test observe exactly which
// userID parameter reaches the service layer (for UpdateAuctionRequest's
// user-id-spoofing regression tests below) and control whether the call succeeds.
func (f *fakeRequestService) UpdateAuctionRequest(ctx context.Context, id uuid.UUID, userID uuid.UUID, updates *models.AuctionRequest) error {
	f.updateCalled = true
	f.updateReceivedUserID = userID
	return f.updateErr
}

func (f *fakeRequestService) AdminUpdateAuctionRequest(ctx context.Context, id uuid.UUID, updates *models.AuctionRequest, insurancePolicy *string) error {
	f.adminUpdateCalled = true
	return f.adminUpdateErr
}

// auctionRequestsResult and bannerRequestsResult let a test control exactly what
// GetAuctionRequests/GetBannerRequests return, for the response-contract regression
// tests below.
func (f *fakeRequestService) GetAuctionRequests(ctx context.Context, status string, userID *uuid.UUID, dateFrom, dateTo *time.Time, categoryID, locationID *int, minPrice, maxPrice *float64, sortBy, sortOrder string, page, perPage int) ([]models.AuctionRequest, int, error) {
	return f.auctionRequestsResult, f.auctionRequestsTotal, f.auctionRequestsErr
}

func (f *fakeRequestService) GetBannerRequests(ctx context.Context, status string, userID *uuid.UUID, dateFrom, dateTo *time.Time, sortBy, sortOrder string, page, perPage int) ([]models.BannerRequest, int, error) {
	return f.bannerRequestsResult, f.bannerRequestsTotal, f.bannerRequestsErr
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

	// No user_id in the body -- models.AuctionRequest.UserID still carries
	// validate:"required" (see request.go), but the handler now assigns the
	// authenticated UserID (see request_handler.go, CreateAuctionRequest) before
	// validate.Struct runs, so a normal documented client payload succeeds without
	// ever having to guess/send its own user id.
	body := map[string]interface{}{
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

// validAuctionRequestBody returns a well-formed create-auction-request JSON body
// with no user_id field -- the documented client contract.
func validAuctionRequestBody() map[string]interface{} {
	return map[string]interface{}{
		"category_id":      4,
		"title_ar":         "اختبار Staging",
		"description_ar":   "وصف تجريبي آمن يزيد عن عشرة أحرف بدون سكريبت",
		"start_price":      100,
		"min_increment":    10,
		"insurance_amount": 0,
		"start_date":       time.Now().Format(time.RFC3339),
		"end_date":         time.Now().Add(24 * time.Hour).Format(time.RFC3339),
	}
}

// validBannerRequestBody mirrors what a fixed mobile client sends: RFC3339
// (UTC, "Z"-suffixed) start/end times and target_url (not link_url) -- see
// Bug A regression tests below.
func validBannerRequestBody() map[string]interface{} {
	return map[string]interface{}{
		"title_ar":       "إعلان اختبار Staging",
		"title_fr":       "Annonce test Staging",
		"title_en":       "Staging test ad",
		"description_ar": "وصف الإعلان",
		"description_fr": "Description",
		"description_en": "Description",
		"image_url":      "https://example.com/banner.jpg",
		"target_url":     "https://example.com",
		"starts_at":      time.Now().UTC().Format(time.RFC3339),
		"ends_at":        time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339),
	}
}

func mustMarshal(t *testing.T, v interface{}) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("failed to marshal request body: %v", err)
	}
	return b
}

// TestCreateAuctionRequest_UserID_A_ValidPayloadWithoutUserIDSucceeds is regression
// case A: an authenticated request with a valid, documented payload that omits
// user_id entirely must not be rejected with a 400 for a missing UserID, and must
// reach the service layer.
func TestCreateAuctionRequest_UserID_A_ValidPayloadWithoutUserIDSucceeds(t *testing.T) {
	fakeSvc := &fakeRequestService{}
	h := NewRequestHandler(fakeSvc, zap.NewNop())

	app := fiber.New()
	userID := uuid.New()
	app.Post("/v1/api/requests/auctions", func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		return h.CreateAuctionRequest(c)
	})

	payload := mustMarshal(t, validAuctionRequestBody())
	req := httptest.NewRequest(http.MethodPost, "/v1/api/requests/auctions", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected HTTP 200 for a valid payload without user_id, got HTTP %d, body: %s", resp.StatusCode, body)
	}
	if !fakeSvc.createCalled {
		t.Fatal("expected the request to reach the service layer, but it did not")
	}
	if fakeSvc.receivedUserID != userID {
		t.Fatalf("expected service to receive the authenticated user id %s, got %s", userID, fakeSvc.receivedUserID)
	}
}

// TestCreateAuctionRequest_UserID_B_JWTOverridesSpoofedBodyUserID is the mandatory
// security regression case B: if the JWT/context authenticated user is A but the
// request body names a different user_id (B), the service must receive A -- never
// B. This proves a malicious client cannot create a request on another user's
// behalf by supplying a different user_id in the body.
func TestCreateAuctionRequest_UserID_B_JWTOverridesSpoofedBodyUserID(t *testing.T) {
	fakeSvc := &fakeRequestService{}
	h := NewRequestHandler(fakeSvc, zap.NewNop())

	app := fiber.New()
	authenticatedUserID := uuid.New() // "A"
	spoofedUserID := uuid.New()       // "B" -- attacker-supplied, must be ignored
	app.Post("/v1/api/requests/auctions", func(c *fiber.Ctx) error {
		c.Locals("user_id", authenticatedUserID)
		return h.CreateAuctionRequest(c)
	})

	body := validAuctionRequestBody()
	body["user_id"] = spoofedUserID.String()
	payload := mustMarshal(t, body)

	req := httptest.NewRequest(http.MethodPost, "/v1/api/requests/auctions", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected HTTP 200, got HTTP %d, body: %s", resp.StatusCode, body)
	}
	if !fakeSvc.createCalled {
		t.Fatal("expected the request to reach the service layer, but it did not")
	}
	if fakeSvc.receivedUserID != authenticatedUserID {
		t.Fatalf("SECURITY REGRESSION: service received user_id %s, expected the authenticated user %s (spoofed body value %s must be ignored)",
			fakeSvc.receivedUserID, authenticatedUserID, spoofedUserID)
	}
}

// TestCreateAuctionRequest_UserID_C_UnauthenticatedRequestRejected is regression
// case C: with no authenticated user in context, the request must be rejected
// (401) and must never reach the service layer.
func TestCreateAuctionRequest_UserID_C_UnauthenticatedRequestRejected(t *testing.T) {
	fakeSvc := &fakeRequestService{}
	h := NewRequestHandler(fakeSvc, zap.NewNop())

	app := fiber.New()
	app.Post("/v1/api/requests/auctions", func(c *fiber.Ctx) error {
		// Deliberately not setting c.Locals("user_id", ...) to simulate a request
		// that never passed authentication middleware.
		return h.CreateAuctionRequest(c)
	})

	payload := mustMarshal(t, validAuctionRequestBody())
	req := httptest.NewRequest(http.MethodPost, "/v1/api/requests/auctions", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected HTTP 401 for an unauthenticated request, got HTTP %d, body: %s", resp.StatusCode, body)
	}
	if fakeSvc.createCalled {
		t.Fatal("expected an unauthenticated request to never reach the service layer, but the service was called")
	}
}

// TestCreateAuctionRequest_UserID_D_InvalidDecimalPricesStillRejected is regression
// case D: removing the UserID validate tag must not weaken price validation --
// invalid decimal prices must still return 400 and never reach the service layer.
func TestCreateAuctionRequest_UserID_D_InvalidDecimalPricesStillRejected(t *testing.T) {
	fakeSvc := &fakeRequestService{}
	h := NewRequestHandler(fakeSvc, zap.NewNop())

	app := fiber.New()
	userID := uuid.New()
	app.Post("/v1/api/requests/auctions", func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		return h.CreateAuctionRequest(c)
	})

	body := validAuctionRequestBody()
	body["start_price"] = -5
	payload := mustMarshal(t, body)

	req := httptest.NewRequest(http.MethodPost, "/v1/api/requests/auctions", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected HTTP 400 for a negative start_price, got HTTP %d, body: %s", resp.StatusCode, body)
	}
	if fakeSvc.createCalled {
		t.Fatal("expected an invalid price to be rejected before reaching the service layer, but the service was called")
	}
}

// TestCreateAuctionRequest_UserID_E_ValidDecimalPayloadStillSucceeds is regression
// case E: a fully valid payload (valid decimal prices, no user_id) must still
// succeed after this fix -- no regression against the earlier decimal.Decimal fix.
func TestCreateAuctionRequest_UserID_E_ValidDecimalPayloadStillSucceeds(t *testing.T) {
	fakeSvc := &fakeRequestService{}
	h := NewRequestHandler(fakeSvc, zap.NewNop())

	app := fiber.New()
	userID := uuid.New()
	app.Post("/v1/api/requests/auctions", func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		return h.CreateAuctionRequest(c)
	})

	payload := mustMarshal(t, validAuctionRequestBody())
	req := httptest.NewRequest(http.MethodPost, "/v1/api/requests/auctions", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected HTTP 200, got HTTP %d, body: %s", resp.StatusCode, body)
	}
	if !fakeSvc.createCalled {
		t.Fatal("expected the request to reach the service layer, but it did not")
	}
}

// TestCreateAuctionRequest_UserID_F_TrustedUserIDNeverNilWhenAuthenticated is an
// explicit defense-in-depth assertion: for any authenticated request, the UserID
// that reaches validate.Struct (and therefore the service layer) must never be
// uuid.Nil. models.AuctionRequest.UserID keeps validate:"required" specifically as
// a safety net for this -- if a future change ever removed the
// req.UserID = authenticatedUserID assignment before validation, this tag would
// catch it (as HTTP 400, not a silent uuid.Nil reaching the database) rather than
// relying solely on the handler's assignment ordering.
func TestCreateAuctionRequest_UserID_F_TrustedUserIDNeverNilWhenAuthenticated(t *testing.T) {
	fakeSvc := &fakeRequestService{}
	h := NewRequestHandler(fakeSvc, zap.NewNop())

	app := fiber.New()
	userID := uuid.New()
	app.Post("/v1/api/requests/auctions", func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		return h.CreateAuctionRequest(c)
	})

	payload := mustMarshal(t, validAuctionRequestBody())
	req := httptest.NewRequest(http.MethodPost, "/v1/api/requests/auctions", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected HTTP 200, got HTTP %d, body: %s", resp.StatusCode, body)
	}
	if !fakeSvc.createCalled {
		t.Fatal("expected the request to reach the service layer, but it did not")
	}
	if fakeSvc.receivedUserID == uuid.Nil {
		t.Fatal("SECURITY REGRESSION: the UserID reaching the service layer for an authenticated request is uuid.Nil")
	}
}

// TestGetAuctionRequests_ResponseContract is the regression test for the Staging
// crash "Uncaught TypeError: H?.filter is not a function" on the Admin /requests
// page. Root cause: GetAuctionRequests used to call OK(c, fiber.Map{"data": requests,
// "total": total, ...}), which double-wrapped the array one level too deep
// ({"success":true,"data":{"data":[...], "total":...}}) versus what the frontend
// expected ({"success":true,"data":[...], "meta":{...}}). This test asserts the wire
// JSON shape directly: top-level "data" must be the array itself, pagination fields
// must live under "meta", and no nested "data.data" must exist.
func TestGetAuctionRequests_ResponseContract(t *testing.T) {
	fakeSvc := &fakeRequestService{
		auctionRequestsResult: []models.AuctionRequest{
			{ID: uuid.New(), CategoryID: 4, TitleAr: "طلب اختبار"},
		},
		auctionRequestsTotal: 1,
	}
	h := NewRequestHandler(fakeSvc, zap.NewNop())

	app := fiber.New()
	app.Get("/v1/api/admin/requests/auctions", h.GetAuctionRequests)

	req := httptest.NewRequest(http.MethodGet, "/v1/api/admin/requests/auctions?page=1&per_page=20", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected HTTP 200, got HTTP %d, body: %s", resp.StatusCode, body)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(body, &parsed); err != nil {
		t.Fatalf("failed to parse response JSON: %v, body: %s", err, body)
	}

	if success, ok := parsed["success"].(bool); !ok || !success {
		t.Fatalf(`expected top-level "success": true, got: %v`, parsed["success"])
	}

	dataField, ok := parsed["data"]
	if !ok {
		t.Fatal(`expected top-level "data" field, none found`)
	}
	dataArray, ok := dataField.([]interface{})
	if !ok {
		t.Fatalf(`expected top-level "data" to be a JSON array (matching frontend AuctionRequest[]), got %T: %v`, dataField, dataField)
	}
	if len(dataArray) != 1 {
		t.Fatalf("expected 1 item in data array, got %d", len(dataArray))
	}

	// The exact bug: data must NOT itself contain a nested "data" key.
	if dataObj, ok := dataArray[0].(map[string]interface{}); ok {
		if _, hasNestedData := dataObj["data"]; hasNestedData {
			t.Fatal(`REGRESSION: response "data" array element unexpectedly contains a nested "data" key -- double-wrapping bug is back`)
		}
	}

	meta, ok := parsed["meta"].(map[string]interface{})
	if !ok {
		t.Fatalf(`expected top-level "meta" object with pagination fields, got: %v`, parsed["meta"])
	}
	if total, ok := meta["total"].(float64); !ok || int(total) != 1 {
		t.Fatalf(`expected meta.total == 1, got: %v`, meta["total"])
	}
	if _, ok := meta["page"]; !ok {
		t.Fatal(`expected meta.page to be present`)
	}
	if _, ok := meta["per_page"]; !ok {
		t.Fatal(`expected meta.per_page to be present`)
	}

	// Confirm no top-level "total"/"page"/"per_page" leaked outside meta (the old,
	// pre-fix shape put them alongside "data" instead of inside "meta").
	for _, leaked := range []string{"total", "page", "per_page"} {
		if _, exists := parsed[leaked]; exists {
			t.Fatalf(`REGRESSION: top-level %q found outside "meta" -- old response shape is back`, leaked)
		}
	}
}

// TestGetBannerRequests_ResponseContract is the identical regression test for
// GET /v1/api/admin/requests/banners, which had the same double-wrapping bug.
func TestUpdateAuctionRequest_A_ValidPayloadWithoutUserIDSucceeds(t *testing.T) {
	fakeSvc := &fakeRequestService{}
	h := NewRequestHandler(fakeSvc, zap.NewNop())

	app := fiber.New()
	userID := uuid.New()
	app.Put("/v1/api/requests/auctions/:id", func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		return h.UpdateAuctionRequest(c)
	})

	payload := mustMarshal(t, validAuctionRequestBody())
	req := httptest.NewRequest(http.MethodPut, "/v1/api/requests/auctions/"+uuid.New().String(), bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected HTTP 200 for a valid payload without user_id, got HTTP %d, body: %s", resp.StatusCode, body)
	}
	if !fakeSvc.updateCalled {
		t.Fatal("expected the request to reach the service layer, but it did not")
	}
	if fakeSvc.updateReceivedUserID != userID {
		t.Fatalf("expected service to receive the authenticated user id %s, got %s", userID, fakeSvc.updateReceivedUserID)
	}
}

func TestUpdateAuctionRequest_B_JWTOverridesSpoofedBodyUserID(t *testing.T) {
	fakeSvc := &fakeRequestService{}
	h := NewRequestHandler(fakeSvc, zap.NewNop())

	app := fiber.New()
	authenticatedUserID := uuid.New()
	spoofedUserID := uuid.New()
	app.Put("/v1/api/requests/auctions/:id", func(c *fiber.Ctx) error {
		c.Locals("user_id", authenticatedUserID)
		return h.UpdateAuctionRequest(c)
	})

	body := validAuctionRequestBody()
	body["user_id"] = spoofedUserID.String()
	payload := mustMarshal(t, body)

	req := httptest.NewRequest(http.MethodPut, "/v1/api/requests/auctions/"+uuid.New().String(), bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected HTTP 200, got HTTP %d, body: %s", resp.StatusCode, body)
	}
	if !fakeSvc.updateCalled {
		t.Fatal("expected the request to reach the service layer, but it did not")
	}
	if fakeSvc.updateReceivedUserID != authenticatedUserID {
		t.Fatalf("SECURITY REGRESSION: service received user_id %s, expected the authenticated user %s (spoofed body value %s must be ignored)",
			fakeSvc.updateReceivedUserID, authenticatedUserID, spoofedUserID)
	}
}

func TestUpdateAuctionRequest_C_UnauthenticatedRequestRejected(t *testing.T) {
	fakeSvc := &fakeRequestService{}
	h := NewRequestHandler(fakeSvc, zap.NewNop())

	app := fiber.New()
	app.Put("/v1/api/requests/auctions/:id", func(c *fiber.Ctx) error {
		return h.UpdateAuctionRequest(c)
	})

	payload := mustMarshal(t, validAuctionRequestBody())
	req := httptest.NewRequest(http.MethodPut, "/v1/api/requests/auctions/"+uuid.New().String(), bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected HTTP 401 for an unauthenticated request, got HTTP %d, body: %s", resp.StatusCode, body)
	}
	if fakeSvc.updateCalled {
		t.Fatal("expected an unauthenticated request to never reach the service layer, but the service was called")
	}
}

func TestUpdateAuctionRequest_D_InvalidDecimalPricesStillRejected(t *testing.T) {
	fakeSvc := &fakeRequestService{}
	h := NewRequestHandler(fakeSvc, zap.NewNop())

	app := fiber.New()
	userID := uuid.New()
	app.Put("/v1/api/requests/auctions/:id", func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		return h.UpdateAuctionRequest(c)
	})

	body := validAuctionRequestBody()
	body["start_price"] = -5
	payload := mustMarshal(t, body)

	req := httptest.NewRequest(http.MethodPut, "/v1/api/requests/auctions/"+uuid.New().String(), bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected HTTP 400 for a negative start_price, got HTTP %d, body: %s", resp.StatusCode, body)
	}
	if fakeSvc.updateCalled {
		t.Fatal("expected an invalid price to be rejected before reaching the service layer, but the service was called")
	}
}

func TestUpdateAuctionRequest_E_RejectedRequestReshapeSucceedsWithoutClientUserID(t *testing.T) {
	fakeSvc := &fakeRequestService{}
	h := NewRequestHandler(fakeSvc, zap.NewNop())

	app := fiber.New()
	userID := uuid.New()
	app.Put("/v1/api/requests/auctions/:id", func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		return h.UpdateAuctionRequest(c)
	})

	body := validAuctionRequestBody()
	body["description_ar"] = "Updated description for resubmit verification testing purposes only"
	payload := mustMarshal(t, body)

	req := httptest.NewRequest(http.MethodPut, "/v1/api/requests/auctions/"+uuid.New().String(), bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected HTTP 200 for a valid resubmit-shaped update, got HTTP %d, body: %s", resp.StatusCode, body)
	}
	if !fakeSvc.updateCalled {
		t.Fatal("expected the request to reach the service layer, but it did not")
	}
}

func TestAdminUpdateAuctionRequest_ValidPayloadWithoutUserIDSucceeds(t *testing.T) {
	fakeSvc := &fakeRequestService{}
	h := NewRequestHandler(fakeSvc, zap.NewNop())

	app := fiber.New()
	adminID := uuid.New()
	app.Put("/v1/api/admin/requests/auctions/:id", func(c *fiber.Ctx) error {
		c.Locals("user_id", adminID)
		return h.AdminUpdateAuctionRequest(c)
	})

	payload := mustMarshal(t, validAuctionRequestBody())
	req := httptest.NewRequest(http.MethodPut, "/v1/api/admin/requests/auctions/"+uuid.New().String(), bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected HTTP 200, got HTTP %d, body: %s", resp.StatusCode, body)
	}
	if !fakeSvc.adminUpdateCalled {
		t.Fatal("expected the request to reach the service layer, but it did not")
	}
}

func TestGetBannerRequests_ResponseContract(t *testing.T) {
	fakeSvc := &fakeRequestService{
		bannerRequestsResult: []models.BannerRequest{
			{ID: uuid.New(), TitleAr: "بانر اختبار"},
		},
		bannerRequestsTotal: 1,
	}
	h := NewRequestHandler(fakeSvc, zap.NewNop())

	app := fiber.New()
	app.Get("/v1/api/admin/requests/banners", h.GetBannerRequests)

	req := httptest.NewRequest(http.MethodGet, "/v1/api/admin/requests/banners?page=1&per_page=20", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected HTTP 200, got HTTP %d, body: %s", resp.StatusCode, body)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(body, &parsed); err != nil {
		t.Fatalf("failed to parse response JSON: %v, body: %s", err, body)
	}

	dataField, ok := parsed["data"]
	if !ok {
		t.Fatal(`expected top-level "data" field, none found`)
	}
	dataArray, ok := dataField.([]interface{})
	if !ok {
		t.Fatalf(`expected top-level "data" to be a JSON array (matching frontend BannerRequest[]), got %T: %v`, dataField, dataField)
	}
	if len(dataArray) != 1 {
		t.Fatalf("expected 1 item in data array, got %d", len(dataArray))
	}

	meta, ok := parsed["meta"].(map[string]interface{})
	if !ok {
		t.Fatalf(`expected top-level "meta" object with pagination fields, got: %v`, parsed["meta"])
	}
	if total, ok := meta["total"].(float64); !ok || int(total) != 1 {
		t.Fatalf(`expected meta.total == 1, got: %v`, meta["total"])
	}

	for _, leaked := range []string{"total", "page", "per_page"} {
		if _, exists := parsed[leaked]; exists {
			t.Fatalf(`REGRESSION: top-level %q found outside "meta" -- old response shape is back`, leaked)
		}
	}
}

// ==================================================
// Bug A (client feedback): the mobile "طلب إعلان" (ad/banner request) page
// returned "Invalid request body" on submit. Root cause, reproduced against
// real Staging: mobile's DateTime.toIso8601String() (no .toUtc()) produced a
// timezone-less string ("2026-09-13T00:00:00.000") that Go's time.Time JSON
// unmarshal rejects, so c.BodyParser failed and the handler returned the
// literal "Invalid request body" -- request_handler.go CreateBannerRequest.
// A second, independent bug was found while proving the date fix alone: with
// the date fixed, models.BannerRequest.UserID (validate:"required") was
// still validated BEFORE being set from the JWT (mirrors the exact bug
// already fixed for CreateAuctionRequest -- see the UserID_A-F tests above),
// so validation failed on the zero UUID regardless of a well-formed body.
// Both were fixed together since neither alone makes a real submission
// succeed. A third, non-blocking issue (mobile sent link_url, backend
// persists target_url, so a user-entered link was silently discarded) was
// fixed on the mobile side since target_url is backend's existing,
// consistent field name.
// ==================================================

// (A-1) A well-formed request (matching the FIXED mobile contract: target_url,
// UTC/RFC3339 dates) is accepted and reaches the service layer with the
// authenticated user's ID, never the zero UUID.
func TestCreateBannerRequest_A_ValidPayloadSucceeds(t *testing.T) {
	fakeSvc := &fakeRequestService{}
	h := NewRequestHandler(fakeSvc, zap.NewNop())

	app := fiber.New()
	userID := uuid.New()
	app.Post("/v1/api/requests/banners", func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		return h.CreateBannerRequest(c)
	})

	payload := mustMarshal(t, validBannerRequestBody())
	req := httptest.NewRequest(http.MethodPost, "/v1/api/requests/banners", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected HTTP 200 for a valid payload, got HTTP %d, body: %s", resp.StatusCode, body)
	}
	if !fakeSvc.createBannerCalled {
		t.Fatal("expected the request to reach the service layer, but it did not")
	}
	if fakeSvc.receivedBannerUserID != userID {
		t.Fatalf("expected service to receive the authenticated user id %s, got %s (this is exactly the UserID-validated-before-set bug)", userID, fakeSvc.receivedBannerUserID)
	}
}

// (A-2) The exact real-device failure, reproduced at the HTTP layer: a
// mobile-shaped payload using DateTime.toIso8601String() with NO timezone
// suffix (the pre-fix mobile behavior) must fail BodyParser with literally
// "Invalid request body" -- proving this is really what real users saw, and
// guarding against ever reintroducing a timezone-less date on this path.
func TestCreateBannerRequest_A2_TimezonelessDate_StillFailsWithSameMessage(t *testing.T) {
	fakeSvc := &fakeRequestService{}
	h := NewRequestHandler(fakeSvc, zap.NewNop())

	app := fiber.New()
	userID := uuid.New()
	app.Post("/v1/api/requests/banners", func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		return h.CreateBannerRequest(c)
	})

	body := validBannerRequestBody()
	// The exact pre-fix mobile shape: Dart's DateTime.toIso8601String() on a
	// local (non-UTC) DateTime, no "Z"/offset suffix.
	body["starts_at"] = "2026-09-13T00:00:00.000"
	body["ends_at"] = "2026-09-20T00:00:00.000"
	payload := mustMarshal(t, body)

	req := httptest.NewRequest(http.MethodPost, "/v1/api/requests/banners", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected HTTP 400 for a timezone-less date, got HTTP %d, body: %s", resp.StatusCode, respBody)
	}
	respBody, _ := io.ReadAll(resp.Body)
	var parsed map[string]interface{}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		t.Fatalf("failed to parse error response: %v", err)
	}
	errObj, _ := parsed["error"].(map[string]interface{})
	if msg, _ := errObj["message"].(string); msg != "Invalid request body" {
		t.Fatalf(`expected the exact real-device error message "Invalid request body", got: %v`, errObj)
	}
	if fakeSvc.createBannerCalled {
		t.Fatal("expected a body-parse failure to never reach the service layer")
	}
}

// (A-3) An unauthenticated request must be rejected (401) and never reach
// the service layer -- mirrors CreateAuctionRequest_UserID_C.
func TestCreateBannerRequest_A3_UnauthenticatedRequestRejected(t *testing.T) {
	fakeSvc := &fakeRequestService{}
	h := NewRequestHandler(fakeSvc, zap.NewNop())

	app := fiber.New()
	app.Post("/v1/api/requests/banners", func(c *fiber.Ctx) error {
		return h.CreateBannerRequest(c)
	})

	payload := mustMarshal(t, validBannerRequestBody())
	req := httptest.NewRequest(http.MethodPost, "/v1/api/requests/banners", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected HTTP 401 for an unauthenticated request, got HTTP %d, body: %s", resp.StatusCode, body)
	}
	if fakeSvc.createBannerCalled {
		t.Fatal("expected an unauthenticated request to never reach the service layer, but the service was called")
	}
}

// (A-4) JWT-authenticated user must always override any user_id present in
// the body -- mirrors CreateAuctionRequest_UserID_B. Not the bug being
// fixed, but the ordering fix (validate after assigning UserID) must not
// weaken this existing guarantee.
func TestCreateBannerRequest_A4_JWTOverridesSpoofedBodyUserID(t *testing.T) {
	fakeSvc := &fakeRequestService{}
	h := NewRequestHandler(fakeSvc, zap.NewNop())

	app := fiber.New()
	authenticatedUserID := uuid.New()
	spoofedUserID := uuid.New()
	app.Post("/v1/api/requests/banners", func(c *fiber.Ctx) error {
		c.Locals("user_id", authenticatedUserID)
		return h.CreateBannerRequest(c)
	})

	body := validBannerRequestBody()
	body["user_id"] = spoofedUserID.String()
	payload := mustMarshal(t, body)

	req := httptest.NewRequest(http.MethodPost, "/v1/api/requests/banners", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected HTTP 200, got HTTP %d, body: %s", resp.StatusCode, body)
	}
	if fakeSvc.receivedBannerUserID != authenticatedUserID {
		t.Fatalf("SECURITY REGRESSION: service received user_id %s, expected the authenticated user %s (spoofed body value %s must be ignored)",
			fakeSvc.receivedBannerUserID, authenticatedUserID, spoofedUserID)
	}
}

// (A-5) target_url is optional: omitting it must still succeed.
func TestCreateBannerRequest_A5_OptionalTargetURLOmitted_StillSucceeds(t *testing.T) {
	fakeSvc := &fakeRequestService{}
	h := NewRequestHandler(fakeSvc, zap.NewNop())

	app := fiber.New()
	userID := uuid.New()
	app.Post("/v1/api/requests/banners", func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		return h.CreateBannerRequest(c)
	})

	body := validBannerRequestBody()
	delete(body, "target_url")
	payload := mustMarshal(t, body)

	req := httptest.NewRequest(http.MethodPost, "/v1/api/requests/banners", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected HTTP 200 with target_url omitted, got HTTP %d, body: %s", resp.StatusCode, respBody)
	}
	if fakeSvc.receivedBannerReq == nil || fakeSvc.receivedBannerReq.TargetURL != nil {
		t.Fatalf("expected TargetURL to be nil when omitted, got: %v", fakeSvc.receivedBannerReq)
	}
}

// (A-6) target_url supplied must reach the service layer in the correct
// field -- proving the mobile-side rename from link_url actually lands in
// the field the backend persists, not silently dropped.
func TestCreateBannerRequest_A6_TargetURLSupplied_ReachesCorrectField(t *testing.T) {
	fakeSvc := &fakeRequestService{}
	h := NewRequestHandler(fakeSvc, zap.NewNop())

	app := fiber.New()
	userID := uuid.New()
	app.Post("/v1/api/requests/banners", func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		return h.CreateBannerRequest(c)
	})

	body := validBannerRequestBody()
	body["target_url"] = "https://example.com/promo"
	payload := mustMarshal(t, body)

	req := httptest.NewRequest(http.MethodPost, "/v1/api/requests/banners", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected HTTP 200, got HTTP %d, body: %s", resp.StatusCode, respBody)
	}
	if fakeSvc.receivedBannerReq == nil || fakeSvc.receivedBannerReq.TargetURL == nil || *fakeSvc.receivedBannerReq.TargetURL != "https://example.com/promo" {
		t.Fatalf("expected TargetURL to be the supplied link, got: %v", fakeSvc.receivedBannerReq)
	}
}

// (A-7) A malformed image_url (not a URL) must still be rejected by
// validation -- proving the UserID-ordering fix didn't weaken other field
// validation (mirrors CreateAuctionRequest_UserID_D).
func TestCreateBannerRequest_A7_InvalidImageURL_StillRejected(t *testing.T) {
	fakeSvc := &fakeRequestService{}
	h := NewRequestHandler(fakeSvc, zap.NewNop())

	app := fiber.New()
	userID := uuid.New()
	app.Post("/v1/api/requests/banners", func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		return h.CreateBannerRequest(c)
	})

	body := validBannerRequestBody()
	body["image_url"] = "not-a-valid-url"
	payload := mustMarshal(t, body)

	req := httptest.NewRequest(http.MethodPost, "/v1/api/requests/banners", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected HTTP 400 for an invalid image_url, got HTTP %d, body: %s", resp.StatusCode, respBody)
	}
	if fakeSvc.createBannerCalled {
		t.Fatal("expected invalid image_url to never reach the service layer")
	}
}

// (A-8) end <= start must still be rejected -- confirms the pre-existing
// business rule in RequestService.CreateBannerRequest (EndsAt.Before(StartsAt))
// is unaffected by these changes. Handler-level test only proves the request
// reaches the service; the business rule itself lives in the service layer
// and is covered by the ends_at/starts_at ordering already, so this is a
// lightweight smoke check that a same-instant window is passed through
// untouched (not rejected at the handler/validation layer, which is correct
// -- only the service enforces ordering).
func TestCreateBannerRequest_A8_EqualStartEnd_ReachesServiceForBusinessRuleCheck(t *testing.T) {
	fakeSvc := &fakeRequestService{}
	h := NewRequestHandler(fakeSvc, zap.NewNop())

	app := fiber.New()
	userID := uuid.New()
	app.Post("/v1/api/requests/banners", func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		return h.CreateBannerRequest(c)
	})

	body := validBannerRequestBody()
	same := time.Now().UTC().Format(time.RFC3339)
	body["starts_at"] = same
	body["ends_at"] = same
	payload := mustMarshal(t, body)

	req := httptest.NewRequest(http.MethodPost, "/v1/api/requests/banners", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected HTTP 200 (handler/validation layer does not enforce start<end, only the service does), got HTTP %d, body: %s", resp.StatusCode, respBody)
	}
	if !fakeSvc.createBannerCalled {
		t.Fatal("expected the request to reach the service layer so its own start<end business rule can run")
	}
}
