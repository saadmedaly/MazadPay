-- Customer Request #34: withdrawal/insurance-refund requests need to record
-- which phone/account number the money should be sent to (e.g. via Bankily
-- "خدمة بنكية"). No existing column on transactions is dedicated to this --
-- payment_method/description are both unused by the withdraw path and would
-- be semantically confusing to overload. Nullable: existing withdraw rows,
-- and any future non-beneficiary-based gateway, are unaffected.
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS beneficiary_account VARCHAR(50);
