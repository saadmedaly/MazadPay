-- Migration rollback: Revert Twilio back to Termii
-- Date: 2025-04-21
-- Description: Rollback migration to restore Termii column

-- Step 1: Add back Termii column
ALTER TABLE otp_verifications ADD COLUMN termii_pin_id VARCHAR(100);

-- Step 2: Copy data from twilio_sid to termii_pin_id
UPDATE otp_verifications SET termii_pin_id = twilio_sid WHERE twilio_sid IS NOT NULL;

-- Step 3: Make termii_pin_id NOT NULL after data migration
ALTER TABLE otp_verifications ALTER COLUMN termii_pin_id SET NOT NULL;

-- Step 4: Drop Twilio column
ALTER TABLE otp_verifications DROP COLUMN twilio_sid;

-- Step 5: Drop Twilio index
DROP INDEX IF EXISTS idx_otp_verifications_twilio_sid;
