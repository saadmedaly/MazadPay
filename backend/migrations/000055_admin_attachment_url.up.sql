-- Customer Request #36: admin transaction review (approve/reject) may
-- optionally attach a review image (client reference: file-upload slot next
-- to the admin notes textarea). Distinct from receipt_url (the USER's
-- uploaded deposit proof) -- this is the ADMIN's own attachment, uploaded
-- via a separate admin-only upload endpoint (no ownership coupling to the
-- transaction, unlike the user-side receipt flow). Nullable: most reviews
-- attach nothing.
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS admin_attachment_url TEXT;
