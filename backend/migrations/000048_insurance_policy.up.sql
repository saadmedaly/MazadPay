-- Insurance policy (client feedback follow-up): distinguishes an explicit
-- admin decision "no insurance required" from the accidental/default
-- insurance_amount = 0 state that caused the V03 incident. Two states only,
-- no 'undecided' -- DEFAULT 'required' means every existing row (and every
-- future row that omits the field) stays under today's exact protection
-- unless an admin explicitly flips it.
ALTER TABLE auction_requests ADD COLUMN insurance_policy VARCHAR(20) NOT NULL DEFAULT 'required';
ALTER TABLE auction_requests ADD CONSTRAINT chk_ar_insurance_policy
    CHECK (insurance_policy IN ('required', 'not_required'));

ALTER TABLE auctions ADD COLUMN insurance_policy VARCHAR(20) NOT NULL DEFAULT 'required';
ALTER TABLE auctions ADD CONSTRAINT chk_auctions_insurance_policy
    CHECK (insurance_policy IN ('required', 'not_required'));
