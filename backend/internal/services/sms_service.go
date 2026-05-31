package services

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	apperr "github.com/mazadpay/backend/internal/errors"
	"go.uber.org/zap"
)

// Wablas API endpoint path used for sending OTP messages.
const wablasSendPath = "/api/send-message"

type SMSService interface {
	SendOTP(phone, code string) error
	IsConfigured() bool
}

type smsService struct {
	token     string
	secretKey string
	serverURL string
	logger    *zap.Logger
}

func NewSMSService(token, secretKey, serverURL string, logger *zap.Logger) SMSService {
	return &smsService{
		token:     token,
		secretKey: secretKey,
		serverURL: serverURL,
		logger:    logger,
	}
}

func (s *smsService) IsConfigured() bool {
	return s.token != "" && s.secretKey != ""
}

func (s *smsService) SendOTP(phone, code string) error {
	if s.token == "" || s.secretKey == "" {
		s.logger.Error("Wablas is not configured: missing token or secret key")
		return apperr.ErrWablasNotConfigured
	}

	cleanPhone := strings.TrimPrefix(phone, "+")
	message := fmt.Sprintf("رمز التحقق الخاص بك هو: %s", code)

	apiURL := s.serverURL + wablasSendPath
	formData := url.Values{
		"phone":   {cleanPhone},
		"message": {message},
		"flag":    {"instant"},
	}

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(formData.Encode()))
	if err != nil {
		s.logger.Error("failed to create Wablas request", zap.Error(err))
		return fmt.Errorf("failed to send WhatsApp OTP: %w", err)
	}

	auth := fmt.Sprintf("%s.%s", s.token, s.secretKey)
	req.Header.Set("Authorization", auth)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		s.logger.Error("failed to send WhatsApp OTP", zap.Error(err), zap.String("phone", phone))
		return fmt.Errorf("failed to send WhatsApp OTP: %w", err)
	}
	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		s.logger.Error("failed to read Wablas response body", zap.Error(readErr))
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		s.logger.Error("Wablas API returned non-2xx status",
			zap.Int("status", resp.StatusCode),
			zap.String("body", string(body)),
		)
		return fmt.Errorf("failed to send WhatsApp OTP: status %d", resp.StatusCode)
	}

	s.logger.Info("WhatsApp OTP sent successfully",
		zap.String("phone", phone),
		zap.Int("http_status", resp.StatusCode),
	)
	return nil
}

func GenerateOTP(length int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	code := ""
	for i := 0; i < length; i++ {
		code += fmt.Sprintf("%d", r.Intn(10))
	}
	return code
}

// ValidatePINStrength vérifie que le PIN respecte les critères de sécurité minimum
// Critères:
// - 4 chiffres minimum
// - Ne doit pas être répétitif (1111, 2222, etc)
// - Ne doit pas être séquentiel simple (1234, 4321, etc)
func ValidatePINStrength(pin string) error {
	if len(pin) < 4 {
		return apperr.ErrWeakPin
	}

	// Vérifier qu'il ne soit pas tout identique (1111, 2222, etc)
	allSame := true
	for i := 1; i < len(pin); i++ {
		if pin[i] != pin[0] {
			allSame = false
			break
		}
	}
	if allSame {
		return apperr.ErrWeakPin
	}

	// Vérifier qu'il ne soit pas séquentiel (1234, 4321, etc)
	if isSequential(pin) {
		return apperr.ErrWeakPin
	}

	return nil
}

func isSequential(s string) bool {
	if len(s) < 2 {
		return false
	}

	asc := true
	desc := true

	for i := 1; i < len(s); i++ {
		diff := int(s[i]) - int(s[i-1])
		if diff != 1 {
			asc = false
		}
		if diff != -1 {
			desc = false
		}
	}

	return asc || desc
}
