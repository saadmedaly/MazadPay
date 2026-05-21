-- Migration: Replace Termii with Twilio in OTP system
-- Date: 2025-04-21
-- Description: Update OTP verification system to use Twilio instead of Termii
-- Note: Made idempotent with IF NOT EXISTS / IF EXISTS guards.

-- Step 1: Add Twilio SID column if not already present
ALTER TABLE otp_verifications ADD COLUMN IF NOT EXISTS twilio_sid VARCHAR(100);

-- Step 2: Copy data from termii_pin_id to twilio_sid if that column exists
-- (safe no-op if termii_pin_id was never added)
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name='otp_verifications' AND column_name='termii_pin_id'
    ) THEN
        UPDATE otp_verifications SET twilio_sid = termii_pin_id WHERE termii_pin_id IS NOT NULL;
    END IF;
END $$;

-- Step 3: Make twilio_sid NOT NULL only if it is currently nullable
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name='otp_verifications' AND column_name='twilio_sid' AND is_nullable='YES'
    ) THEN
        UPDATE otp_verifications SET twilio_sid = 'legacy_' || id::text WHERE twilio_sid IS NULL OR twilio_sid = '';
        ALTER TABLE otp_verifications ALTER COLUMN twilio_sid SET NOT NULL;
    END IF;
END $$;

-- Step 4: Drop termii_pin_id if it still exists
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name='otp_verifications' AND column_name='termii_pin_id'
    ) THEN
        ALTER TABLE otp_verifications DROP COLUMN termii_pin_id;
    END IF;
END $$;

-- Step 5: Create index if not exists
CREATE INDEX IF NOT EXISTS idx_otp_verifications_twilio_sid ON otp_verifications(twilio_sid);
