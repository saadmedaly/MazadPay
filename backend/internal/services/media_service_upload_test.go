package services

import (
	"bytes"
	"mime/multipart"
	"net/textproto"
	"os"
	"testing"

	"github.com/mazadpay/backend/internal/config"
	"go.uber.org/zap"
)

// MAZADPAY -- withdrawal approval + payment receipt bug: the admin panel's
// "upload transfer proof" flow rejected genuine PNG/JPG screenshots with
// "file content does not match a valid .png file". Root cause: MediaService's
// validateUpload required an EXACT 1:1 match between the client-supplied
// filename extension and the magic-byte-detected content type (".png" MUST
// detect as EXACTLY "image/png", nothing else) -- but real screenshots from
// messaging apps/OS tools routinely get re-encoded or renamed without their
// extension staying in sync with their real bytes, so a perfectly valid,
// safe image could be rejected purely because its filename lied about its
// own format. Fixed by accepting ANY real image content type
// (jpg/jpeg/png/webp/gif) for ANY image extension, while still fully
// rejecting non-image content (the actual security property this validation
// exists for).
//
// These tests exercise the REAL MediaService.UploadFile (local-storage mode,
// no R2 credentials in test config -- see isR2Configured), not a mock, so
// they prove the actual magic-byte validation logic.

func newTestMediaService(t *testing.T) *mediaService {
	t.Helper()
	tmpDir := t.TempDir()
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd failed: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Chdir failed: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(origWd) })

	cfg := &config.Config{App: config.AppConfig{Env: "test", Port: "8082"}}
	svc := NewMediaService(cfg, zap.NewNop())
	ms, ok := svc.(*mediaService)
	if !ok {
		t.Fatalf("expected *mediaService, got %T", svc)
	}
	if !ms.useLocal {
		t.Fatalf("expected local-storage fallback in test (no R2 configured)")
	}
	return ms
}

// realPNGBytes are the genuine PNG magic-byte header + minimal valid IHDR
// chunk -- a real, safe image, not a renamed non-image file.
var realPNGBytes = []byte{
	0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, // PNG signature
	0x00, 0x00, 0x00, 0x0D, 'I', 'H', 'D', 'R',
	0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
	0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4, 0x89,
}

// realJPEGBytes are the genuine JPEG magic-byte header.
var realJPEGBytes = []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 'J', 'F', 'I', 'F', 0x00, 0x01}

func buildMultipartFile(t *testing.T, filename string, content []byte) (multipart.File, *multipart.FileHeader) {
	t.Helper()
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreatePart(textproto.MIMEHeader{
		"Content-Disposition": []string{`form-data; name="file"; filename="` + filename + `"`},
	})
	if err != nil {
		t.Fatalf("CreatePart failed: %v", err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("writer.Close failed: %v", err)
	}

	reader := multipart.NewReader(&buf, writer.Boundary())
	form, err := reader.ReadForm(10 << 20)
	if err != nil {
		t.Fatalf("ReadForm failed: %v", err)
	}
	t.Cleanup(func() { _ = form.RemoveAll() })

	header := form.File["file"][0]
	f, err := header.Open()
	if err != nil {
		t.Fatalf("header.Open failed: %v", err)
	}
	t.Cleanup(func() { _ = f.Close() })
	return f, header
}

func TestUploadFile_ValidPNG_Accepted(t *testing.T) {
	ms := newTestMediaService(t)
	f, header := buildMultipartFile(t, "receipt.png", realPNGBytes)

	url, err := ms.UploadFile(t.Context(), f, header, "transaction-review")
	if err != nil {
		t.Fatalf("expected a real PNG (.png extension, real PNG bytes) to be accepted, got: %v", err)
	}
	if url == "" {
		t.Fatal("expected a non-empty upload URL")
	}
}

func TestUploadFile_ValidJPG_Accepted(t *testing.T) {
	ms := newTestMediaService(t)
	f, header := buildMultipartFile(t, "receipt.jpg", realJPEGBytes)

	url, err := ms.UploadFile(t.Context(), f, header, "transaction-review")
	if err != nil {
		t.Fatalf("expected a real JPG (.jpg extension, real JPEG bytes) to be accepted, got: %v", err)
	}
	if url == "" {
		t.Fatal("expected a non-empty upload URL")
	}
}

func TestUploadFile_ValidJPEG_Accepted(t *testing.T) {
	ms := newTestMediaService(t)
	f, header := buildMultipartFile(t, "receipt.jpeg", realJPEGBytes)

	url, err := ms.UploadFile(t.Context(), f, header, "transaction-review")
	if err != nil {
		t.Fatalf("expected a real JPEG (.jpeg extension, real JPEG bytes) to be accepted, got: %v", err)
	}
	if url == "" {
		t.Fatal("expected a non-empty upload URL")
	}
}

// TestUploadFile_RealScreenshot_MismatchedExtension_Accepted is the exact
// regression this ticket reports: a genuine image whose filename extension
// does not match its real detected content type (e.g. a screenshot
// re-encoded by a messaging app) must still be accepted -- it is real,
// safe image content, not a spoofed file.
func TestUploadFile_RealScreenshot_MismatchedExtension_Accepted(t *testing.T) {
	ms := newTestMediaService(t)
	// Real PNG bytes, but named ".jpg" -- exactly the mismatch a re-encoded
	// WhatsApp/Telegram screenshot can produce.
	f, header := buildMultipartFile(t, "receipt.jpg", realPNGBytes)

	url, err := ms.UploadFile(t.Context(), f, header, "transaction-review")
	if err != nil {
		t.Fatalf("expected a real PNG named .jpg to be accepted (content is genuinely an image), got: %v", err)
	}
	if url == "" {
		t.Fatal("expected a non-empty upload URL")
	}
}

// TestUploadFile_FakeImageTextFile_Rejected proves the security property this
// validation exists for is still fully enforced: a non-image file (plain
// text) renamed to .png must still be rejected.
func TestUploadFile_FakeImageTextFile_Rejected(t *testing.T) {
	ms := newTestMediaService(t)
	f, header := buildMultipartFile(t, "not-a-real-image.png", []byte("this is just plain text, not an image at all"))

	_, err := ms.UploadFile(t.Context(), f, header, "transaction-review")
	if err == nil {
		t.Fatal("expected a fake .png (plain text content) to be rejected")
	}
}

func TestUploadFile_FakeImageExecutable_Rejected(t *testing.T) {
	ms := newTestMediaService(t)
	// MZ header -- a Windows executable renamed to .png.
	f, header := buildMultipartFile(t, "malware.png", []byte{0x4D, 0x5A, 0x90, 0x00, 0x03, 0x00, 0x00, 0x00})

	_, err := ms.UploadFile(t.Context(), f, header, "transaction-review")
	if err == nil {
		t.Fatal("expected a renamed executable (.png extension, MZ header) to be rejected")
	}
}
