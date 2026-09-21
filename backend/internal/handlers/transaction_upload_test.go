package handlers

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/mazadpay/backend/internal/middleware"
	"github.com/mazadpay/backend/internal/services"
	"go.uber.org/zap"
)

// MAZADPAY -- withdrawal approval + payment receipt bug: end-to-end test for
// POST /v1/api/admin/transactions/upload, wired through the REAL
// jwtMiddleware + AdminOnly middleware chain (identical to production
// routing.go registration), with a fake in-memory MediaService whose
// UploadFile re-validates real magic bytes exactly like the real
// MediaService.validateUpload (see media_service_upload_test.go for the
// underlying MediaService-level coverage of the actual fix).

const testTxnUploadJWTSecret = "test-secret-mazadpay-txn-upload"

type fakeMediaServiceForTxnUpload struct {
	uploadCount int
}

func (f *fakeMediaServiceForTxnUpload) UploadFile(ctx context.Context, file multipart.File, header *multipart.FileHeader, folder string) (string, error) {
	f.uploadCount++
	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	contentType := http.DetectContentType(buf[:n])
	if idx := bytes.IndexByte([]byte(contentType), ';'); idx != -1 {
		contentType = contentType[:idx]
	}
	allowed := map[string]bool{"image/jpeg": true, "image/png": true, "image/webp": true, "image/gif": true}
	if !allowed[contentType] {
		return "", fiber.NewError(fiber.StatusBadRequest, "file content does not match a valid image file")
	}
	return "https://cdn.example.com/transaction-review/" + uuid.New().String() + ".png", nil
}
func (f *fakeMediaServiceForTxnUpload) UploadAuctionImages(ctx context.Context, files []multipart.File, headers []*multipart.FileHeader, auctionID uuid.UUID) ([]string, error) {
	return nil, nil
}
func (f *fakeMediaServiceForTxnUpload) DeleteFile(ctx context.Context, key string) error { return nil }
func (f *fakeMediaServiceForTxnUpload) GetPublicURL(key string) string                  { return key }
func (f *fakeMediaServiceForTxnUpload) ExtractKey(url string) string                    { return url }
func (f *fakeMediaServiceForTxnUpload) GetPresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	return key, nil
}
func (f *fakeMediaServiceForTxnUpload) GetReceiptURL(ctx context.Context, storedValue string, expiry time.Duration) (string, error) {
	return storedValue, nil
}
func (f *fakeMediaServiceForTxnUpload) UploadPrivateFile(ctx context.Context, file multipart.File, header *multipart.FileHeader, folder string) (string, error) {
	return "", nil
}

var _ services.MediaService = (*fakeMediaServiceForTxnUpload)(nil)

func buildTxnUploadTestApp(t *testing.T, mediaSvc services.MediaService) *fiber.App {
	t.Helper()
	app := fiber.New()
	logger := zap.NewNop()
	adminHandler := NewAdminHandler(&fakeAdminServiceForValidateTxn{}, nil, logger)

	jwtMW := middleware.JWT(testTxnUploadJWTSecret, logger, nil)
	adminMW := middleware.AdminOnly(logger)

	admin := app.Group("/v1/api/admin/transactions", jwtMW, adminMW)
	admin.Post("/upload", func(c *fiber.Ctx) error {
		c.Locals("mediaService", mediaSvc)
		return adminHandler.UploadTransactionReviewAttachment(c)
	})
	return app
}

func signTxnUploadTestJWT(t *testing.T) string {
	t.Helper()
	claims := services.JWTClaims{
		UserID: uuid.New().String(),
		Role:   "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(testTxnUploadJWTSecret))
	if err != nil {
		t.Fatalf("failed to sign test JWT: %v", err)
	}
	return signed
}

func buildMultipartUploadRequest(t *testing.T, filename string, content []byte, token string) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("CreateFormFile failed: %v", err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("writer.Close failed: %v", err)
	}

	req := httptest.NewRequest("POST", "/v1/api/admin/transactions/upload", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return req
}

func TestUploadTransactionReviewAttachment_ValidPNG_Accepted(t *testing.T) {
	mediaSvc := &fakeMediaServiceForTxnUpload{}
	app := buildTxnUploadTestApp(t, mediaSvc)
	token := signTxnUploadTestJWT(t)

	req := buildMultipartUploadRequest(t, "receipt.png", realPNGBytesForHandlerTest, token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 for a valid PNG, got %d", resp.StatusCode)
	}
	if mediaSvc.uploadCount != 1 {
		t.Fatalf("expected exactly 1 upload call, got %d", mediaSvc.uploadCount)
	}
}

func TestUploadTransactionReviewAttachment_FakeImage_Rejected(t *testing.T) {
	mediaSvc := &fakeMediaServiceForTxnUpload{}
	app := buildTxnUploadTestApp(t, mediaSvc)
	token := signTxnUploadTestJWT(t)

	req := buildMultipartUploadRequest(t, "not-real.png", []byte("just plain text pretending to be a png"), token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	if resp.StatusCode == fiber.StatusOK {
		t.Fatal("expected a fake .png (plain text content) to be rejected, got 200")
	}
}

// TestUploadTransactionReviewAttachment_UppercaseExtension_Accepted: the
// handler's own extension whitelist check (separate from MediaService's
// magic-byte check) was case-sensitive, so a Windows screenshot tool's
// "Screenshot.PNG" (uppercase extension) would be wrongly rejected before
// ever reaching the real content validation.
func TestUploadTransactionReviewAttachment_UppercaseExtension_Accepted(t *testing.T) {
	mediaSvc := &fakeMediaServiceForTxnUpload{}
	app := buildTxnUploadTestApp(t, mediaSvc)
	token := signTxnUploadTestJWT(t)

	req := buildMultipartUploadRequest(t, "Screenshot.PNG", realPNGBytesForHandlerTest, token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 for a valid PNG with an uppercase .PNG extension, got %d", resp.StatusCode)
	}
}

var realPNGBytesForHandlerTest = []byte{
	0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
	0x00, 0x00, 0x00, 0x0D, 'I', 'H', 'D', 'R',
	0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
	0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4, 0x89,
}
