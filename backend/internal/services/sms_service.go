package services

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/twilio/twilio-go"
	twilioApi "github.com/twilio/twilio-go/rest/api/v2010"
	"go.uber.org/zap"
)

type SMSService interface {
	SendOTP(phone, code string) error
}

type smsService struct {
	accountSID   string
	authToken   string
	fromNumber string
	logger    *zap.Logger
}

func NewSMSService(accountSID, authToken, fromNumber string, logger *zap.Logger) SMSService {
	return &smsService{
		accountSID:   accountSID,
		authToken:   authToken,
		fromNumber: fromNumber,
		logger:     logger,
	}
}

func (s *smsService) SendOTP(phone, code string) error {
	if s.accountSID == "" || s.authToken == "" {
		s.logger.Warn("Twilio not configured, skipping SMS send")
		return nil
	}

	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: s.accountSID,
		Password: s.authToken,
	})

	message := fmt.Sprintf("Your verification code is: %s", code)

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