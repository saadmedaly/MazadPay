-- ============================================================
-- Down migration for 000045_auction_request_draft_status
--
-- NOTE: this will fail (as intended) if any row currently has
-- status = 'draft' at rollback time, since that value would violate the
-- narrower constraint being restored. That is safe, expected behavior for a
-- down-migration -- it should fail loudly rather than silently corrupt or
-- orphan draft rows.
-- ============================================================

ALTER TABLE auction_requests DROP CONSTRAINT auction_requests_status_check;
ALTER TABLE auction_requests ADD CONSTRAINT auction_requests_status_check
    CHECK (status IN ('pending', 'approved', 'rejected'));
