-- Migration: Replace Twilio with Wablas WhatsApp OTP
-- Date: 2026-05-27
-- Description: Rename twilio_sid column to otp_code in otp_verifications table

ALTER TABLE otp_verifications RENAME COLUMN twilio_sid TO otp_code;
DROP INDEX IF EXISTS idx_otp_verifications_twilio_sid;
CREATE INDEX IF NOT EXISTS idx_otp_verifications_otp_code ON otp_verifications(otp_code);
