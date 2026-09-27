interface InsuranceFields {
  insurance_policy?: 'required' | 'not_required'
  insurance_amount?: string | number
}

/**
 * requestIsApprovableForInsurance mirrors the backend's approval gate
 * exactly -- both ReviewAuctionRequest (request_service.go, for
 * auction_requests) and ValidateAuction (admin_service.go, for auctions
 * already sitting at status 'pending' -- a second, separate admin approval
 * surface): ready for approval, insurance-wise, only when either an admin
 * has explicitly set insurance_policy = 'not_required', or insurance_amount
 * is a positive number under the default 'required' policy. Missing/
 * undefined insurance_policy is treated as 'required', never as
 * 'not_required' (matches the backend's InsuranceRequired() fallback).
 *
 * Structurally typed (not tied to AuctionRequest specifically) so it can be
 * shared across every admin surface that gates approval on insurance:
 * RequestDetailModal.tsx + KYCPage.tsx (auction_requests review) and
 * AuctionsPage.tsx (direct auctions validate) -- client feedback
 * requirement: these entry points can never disagree.
 */
export function requestIsApprovableForInsurance(req: InsuranceFields): boolean {
  if (req.insurance_policy === 'not_required') return true
  return Number(req.insurance_amount) > 0
}
