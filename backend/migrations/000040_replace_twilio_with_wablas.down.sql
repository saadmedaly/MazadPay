-- Migration rollback: Revert otp_code back to twilio_sid
-- Date: 2026-05-27
-- Description: Rename otp_code column back to twilio_sid

ALTER TABLE otp_verifications RENAME COLUMN otp_code TO twilio_sid;
DROP INDEX IF EXISTS idx_otp_verifications_otp_code;
CREATE INDEX IF NOT EXISTS idx_otp_verifications_twilio_sid ON otp_verifications(twilio_sid);
