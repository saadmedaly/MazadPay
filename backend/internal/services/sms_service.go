package services

import (
	"fmt"
	"math/rand"
	"time"

	apperr "github.com/mazadpay/backend/internal/errors"
	"github.com/twilio/twilio-go"
	twilioApi "github.com/twilio/twilio-go/rest/api/v2010"
	"go.uber.org/zap"
)

type SMSService interface {
	SendOTP(phone, code string) error
}

type smsService struct {
	accountSID string
	authToken  string
	fromNumber string
	logger     *zap.Logger
}

func NewSMSService(accountSID, authToken, fromNumber string, logger *zap.Logger) SMSService {
	return &smsService{
		accountSID: accountSID,
		authToken:  authToken,
		fromNumber: fromNumber,
		logger:     logger,
	}
}

func (s *smsService) SendOTP(phone, code string) error {
	if s.accountSID == "" || s.authToken == "" || s.fromNumber == "" {
		s.logger.Error("Twilio is not configured")
		return apperr.ErrTwilioNotConfigured
	}

	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: s.accountSID,
		Password: s.authToken,
	})

	message := fmt.Sprintf("رمز التحقق الخاص بك هو: %s", code)

	params := &twilioApi.CreateMessageParams{}
	params.SetTo(phone)
	params.SetFrom(s.fromNumber)
	params.SetBody(message)

	resp, err := client.Api.CreateMessage(params)
	if err != nil {
		s.logger.Error("failed to send SMS", zap.Error(err), zap.String("phone", phone))
		return fmt.Errorf("failed to send SMS: %w", err)
	}

	s.logger.Info("SMS sent successfully", zap.String("sid", *resp.Sid), zap.String("phone", phone))
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
