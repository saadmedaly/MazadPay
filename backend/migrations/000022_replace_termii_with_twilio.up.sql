-- Migration: Replace Termii with Twilio in OTP system
-- Date: 2025-04-21
-- Description: Update OTP verification system to use Twilio instead of Termii

-- Step 1: Add temporary column for Twilio SID
ALTER TABLE otp_verifications ADD COLUMN twilio_sid VARCHAR(100);

-- Step 2: Copy data from termii_pin_id to twilio_sid (if any data exists)
UPDATE otp_verifications SET twilio_sid = termii_pin_id WHERE termii_pin_id IS NOT NULL;

-- Step 3: Make twilio_sid NOT NULL after data migration
ALTER TABLE otp_verifications ALTER COLUMN twilio_sid SET NOT NULL;

-- Step 4: Drop the old Termii column
ALTER TABLE otp_verifications DROP COLUMN termii_pin_id;

-- Step 5: Add index for better performance on Twilio SID
CREATE INDEX idx_otp_verifications_twilio_sid ON otp_verifications(twilio_sid);

-- Step 6: Update any existing records that might have empty twilio_sid
-- This is a safety measure - in production, this should not happen
UPDATE otp_verifications 
SET twilio_sid = 'temp_' || id::text 
WHERE twilio_sid = '' OR twilio_sid IS NULL;
