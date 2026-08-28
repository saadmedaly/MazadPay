-- ============================================================
-- Allow a 'draft' status on auction_requests, so a seller can save a
-- listing before submitting it for admin review (in addition to the
-- existing pending -> approved/rejected review flow).
-- ============================================================

-- Postgres auto-names an inline column CHECK as <table>_<column>_check, and
-- migration 000003 declared this constraint inline (status VARCHAR(20)
-- DEFAULT 'pending' CHECK (status IN (...))), so the default-generated name
-- is auction_requests_status_check.
ALTER TABLE auction_requests DROP CONSTRAINT auction_requests_status_check;
ALTER TABLE auction_requests ADD CONSTRAINT auction_requests_status_check
    CHECK (status IN ('draft', 'pending', 'approved', 'rejected'));
