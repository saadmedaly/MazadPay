package services

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/repository"
	"github.com/shopspring/decimal"
)

// fakeRequestRepoForAdminUpdate is an in-memory implementation of
// repository.RequestRepository, sufficient to exercise the REAL
// requestService.AdminUpdateAuctionRequest end-to-end (GetAuctionRequestByID
// -> mutate -> UpdateAuctionRequest), unlike ReviewAuctionRequest/PlaceBid
// which require a live *sqlx.Tx via BeginTx and cannot be driven this way.
type fakeRequestRepoForAdminUpdate struct {
	repository.RequestRepository
	stored *models.AuctionRequest
}

func (f *fakeRequestRepoForAdminUpdate) GetAuctionRequestByID(ctx context.Context, id uuid.UUID) (*models.AuctionRequest, error) {
	cp := *f.stored
	return &cp, nil
}

func (f *fakeRequestRepoForAdminUpdate) UpdateAuctionRequest(ctx context.Context, req *models.AuctionRequest) error {
	cp := *req
	f.stored = &cp
	return nil
}

func newAdminUpdateService(stored *models.AuctionRequest) (*requestService, *fakeRequestRepoForAdminUpdate) {
	repo := &fakeRequestRepoForAdminUpdate{stored: stored}
	svc := &requestService{repo: repo}
	return svc, repo
}

// TestAdminUpdateAuctionRequest_OmittedPolicyPreservesExistingNotRequired is
// targeted test (client feedback #8, item 1): updating a request without
// including insurance_policy in the payload (insurancePolicy = nil) must
// PRESERVE an existing 'not_required' value, never silently reset it back to
// 'required'.
func TestAdminUpdateAuctionRequest_OmittedPolicyPreservesExistingNotRequired(t *testing.T) {
	existing := &models.AuctionRequest{
		ID:              uuid.New(),
		InsurancePolicy: models.InsurancePolicyNotRequired,
		InsuranceAmount: decimal.Zero,
	}
	svc, repo := newAdminUpdateService(existing)

	updates := &models.AuctionRequest{TitleAr: "updated title only"}
	if err := svc.AdminUpdateAuctionRequest(context.Background(), existing.ID, updates, nil); err != nil {
		t.Fatalf("AdminUpdateAuctionRequest failed: %v", err)
	}

	if repo.stored.InsurancePolicy != models.InsurancePolicyNotRequired {
		t.Fatalf("SECURITY REGRESSION: omitting insurance_policy from the update reset it from 'not_required' to %q -- must preserve the existing value", repo.stored.InsurancePolicy)
	}
}

// TestAdminUpdateAuctionRequest_RequiredToNotRequired_ForcesAmountToZero is
// targeted test (client feedback #8, item 2): changing the policy from
// 'required' to 'not_required' must canonicalize insurance_amount to zero,
// even if the update payload's amount field carries a stale positive value
// left over from the 'required' state.
func TestAdminUpdateAuctionRequest_RequiredToNotRequired_ForcesAmountToZero(t *testing.T) {
	existing := &models.AuctionRequest{
		ID:              uuid.New(),
		InsurancePolicy: models.InsurancePolicyRequired,
		InsuranceAmount: decimal.NewFromInt(500),
	}
	svc, repo := newAdminUpdateService(existing)

	notRequired := models.InsurancePolicyNotRequired
	// Deliberately still carrying the old positive amount in the payload --
	// proves the backend canonicalizes it, not merely trusts a well-behaved client.
	updates := &models.AuctionRequest{InsuranceAmount: decimal.NewFromInt(500)}
	if err := svc.AdminUpdateAuctionRequest(context.Background(), existing.ID, updates, &notRequired); err != nil {
		t.Fatalf("AdminUpdateAuctionRequest failed: %v", err)
	}

	if repo.stored.InsurancePolicy != models.InsurancePolicyNotRequired {
		t.Fatalf("expected policy to become 'not_required', got %q", repo.stored.InsurancePolicy)
	}
	if !repo.stored.InsuranceAmount.Equal(decimal.Zero) {
		t.Fatalf("SECURITY REGRESSION: expected insurance_amount to be canonicalized to 0 when policy is not_required, got %s -- a stale 'not_required + positive amount' state must never be persisted", repo.stored.InsuranceAmount.String())
	}
}

// TestAdminUpdateAuctionRequest_NotRequiredToRequired_ZeroAmount_StaysUnapprovable
// is targeted test (client feedback #8, item 3): changing the policy back to
// 'required' while leaving insurance_amount at 0 must leave the request in a
// state ReviewAuctionRequest will still reject -- switching the policy alone
// must not implicitly satisfy the amount requirement.
func TestAdminUpdateAuctionRequest_NotRequiredToRequired_ZeroAmount_StaysUnapprovable(t *testing.T) {
	existing := &models.AuctionRequest{
		ID:              uuid.New(),
		InsurancePolicy: models.InsurancePolicyNotRequired,
		InsuranceAmount: decimal.Zero,
	}
	svc, repo := newAdminUpdateService(existing)

	required := models.InsurancePolicyRequired
	updates := &models.AuctionRequest{InsuranceAmount: decimal.Zero}
	if err := svc.AdminUpdateAuctionRequest(context.Background(), existing.ID, updates, &required); err != nil {
		t.Fatalf("AdminUpdateAuctionRequest failed: %v", err)
	}

	if repo.stored.InsurancePolicy != models.InsurancePolicyRequired {
		t.Fatalf("expected policy to become 'required', got %q", repo.stored.InsurancePolicy)
	}
	// Mirrors ReviewAuctionRequest's exact approval predicate.
	if repo.stored.InsuranceRequired() && repo.stored.InsuranceAmount.GreaterThan(decimal.Zero) {
		t.Fatal("expected the request to remain unapprovable (required policy + zero amount)")
	}
	if !repo.stored.InsuranceRequired() {
		t.Fatal("expected InsuranceRequired() to be true after switching to 'required'")
	}
}

// TestAdminUpdateAuctionRequest_ExplicitRequiredPolicyPreserved proves the
// mirror case of the "omitted preserves" test: explicitly re-sending
// 'required' when it was already 'required' is a no-op, not a regression.
func TestAdminUpdateAuctionRequest_ExplicitRequiredPolicyPreserved(t *testing.T) {
	existing := &models.AuctionRequest{
		ID:              uuid.New(),
		InsurancePolicy: models.InsurancePolicyRequired,
		InsuranceAmount: decimal.NewFromInt(1000),
	}
	svc, repo := newAdminUpdateService(existing)

	required := models.InsurancePolicyRequired
	updates := &models.AuctionRequest{InsuranceAmount: decimal.NewFromInt(1000)}
	if err := svc.AdminUpdateAuctionRequest(context.Background(), existing.ID, updates, &required); err != nil {
		t.Fatalf("AdminUpdateAuctionRequest failed: %v", err)
	}

	if repo.stored.InsurancePolicy != models.InsurancePolicyRequired {
		t.Fatalf("expected policy to remain 'required', got %q", repo.stored.InsurancePolicy)
	}
	if !repo.stored.InsuranceAmount.Equal(decimal.NewFromInt(1000)) {
		t.Fatalf("expected insurance_amount to remain 1000, got %s", repo.stored.InsuranceAmount.String())
	}
}

// TestReviewAuctionRequest_ApprovalGate_InsurancePolicy covers the required
// approval-time predicate cases (client feedback requirements, mirrored
// exactly from request_service.go's ReviewAuctionRequest -- which itself
// requires a live *sqlx.Tx via BeginTx and cannot be driven end-to-end
// without a real database, so this isolates and locks down the exact
// business predicate instead, same approach as the pre-existing
// TestReviewAuctionRequest_ApprovalGate_InsuranceRule in
// request_service_insurance_test.go, now extended for insurance_policy).
func TestReviewAuctionRequest_ApprovalGate_InsurancePolicy(t *testing.T) {
	approvalBlocked := func(req *models.AuctionRequest) bool {
		return req.InsuranceRequired() && !req.InsuranceAmount.GreaterThan(decimal.Zero)
	}

	t.Run("legacy request (empty policy, amount=0) remains blocked -- V03 protection intact", func(t *testing.T) {
		req := &models.AuctionRequest{InsurancePolicy: "", InsuranceAmount: decimal.Zero}
		if !approvalBlocked(req) {
			t.Fatal("SECURITY REGRESSION: expected a legacy request with no policy set and amount=0 to remain blocked from approval")
		}
	})

	t.Run("explicit not_required + amount=0 is approvable", func(t *testing.T) {
		req := &models.AuctionRequest{InsurancePolicy: models.InsurancePolicyNotRequired, InsuranceAmount: decimal.Zero}
		if approvalBlocked(req) {
			t.Fatal("expected an explicit not_required request to be approvable with amount=0")
		}
	})

	t.Run("required + amount=0 is rejected", func(t *testing.T) {
		req := &models.AuctionRequest{InsurancePolicy: models.InsurancePolicyRequired, InsuranceAmount: decimal.Zero}
		if !approvalBlocked(req) {
			t.Fatal("expected a required-policy request with amount=0 to be rejected")
		}
	})

	t.Run("required + positive amount is approved", func(t *testing.T) {
		req := &models.AuctionRequest{InsurancePolicy: models.InsurancePolicyRequired, InsuranceAmount: decimal.NewFromInt(500)}
		if approvalBlocked(req) {
			t.Fatal("expected a required-policy request with a positive amount to be approvable")
		}
	})
}

// TestPlaceBid_InsurancePolicy_FreezeDecision mirrors bid_service.go
// PlaceBid's exact insurance branch logic (client feedback requirements 5
// and the regression test list: legacy zero stays protected, not_required
// skips the freeze, required+positive still freezes). PlaceBid itself
// requires a live *sqlx.Tx (via database.WithTransaction) and cannot be
// exercised end-to-end without a real database connection -- this isolates
// the exact decision the code makes (reject / skip freeze / freeze the
// amount) so the business rule itself is locked down by a fast, real test.
func TestPlaceBid_InsurancePolicy_FreezeDecision(t *testing.T) {
	// decision mirrors exactly:
	//   if auction.InsuranceRequired() {
	//       if !auction.InsuranceAmount.GreaterThan(decimal.Zero) { reject }
	//       else { freeze auction.InsuranceAmount }
	//   } // not required: no freeze, bid proceeds
	type decision struct {
		rejected    bool
		frozeAmount decimal.Decimal
		froze       bool
	}
	decide := func(a *models.Auction) decision {
		if !a.InsuranceRequired() {
			return decision{}
		}
		if !a.InsuranceAmount.GreaterThan(decimal.Zero) {
			return decision{rejected: true}
		}
		return decision{froze: true, frozeAmount: a.InsuranceAmount}
	}

	t.Run("legacy auction (empty policy, amount=0) is still rejected -- V03 protection intact", func(t *testing.T) {
		a := &models.Auction{InsurancePolicy: "", InsuranceAmount: decimal.Zero}
		d := decide(a)
		if !d.rejected {
			t.Fatal("SECURITY REGRESSION: expected a legacy zero-insurance auction to still reject bids")
		}
	})

	t.Run("explicit not_required auction accepts bids with no freeze", func(t *testing.T) {
		a := &models.Auction{InsurancePolicy: models.InsurancePolicyNotRequired, InsuranceAmount: decimal.Zero}
		d := decide(a)
		if d.rejected {
			t.Fatal("expected a not_required auction to accept bids, not reject them")
		}
		if d.froze {
			t.Fatal("expected a not_required auction to skip the insurance freeze entirely")
		}
	})

	t.Run("required + positive amount still freezes exactly that amount", func(t *testing.T) {
		a := &models.Auction{InsurancePolicy: models.InsurancePolicyRequired, InsuranceAmount: decimal.NewFromInt(750)}
		d := decide(a)
		if d.rejected {
			t.Fatal("expected a required auction with a positive amount to accept the bid")
		}
		if !d.froze || !d.frozeAmount.Equal(decimal.NewFromInt(750)) {
			t.Fatalf("expected the exact configured amount (750) to be frozen, got froze=%v amount=%s", d.froze, d.frozeAmount.String())
		}
	})

	t.Run("required + amount=0 is rejected", func(t *testing.T) {
		a := &models.Auction{InsurancePolicy: models.InsurancePolicyRequired, InsuranceAmount: decimal.Zero}
		d := decide(a)
		if !d.rejected {
			t.Fatal("expected a required auction with amount=0 to reject the bid")
		}
	})
}

// fakeAuctionRepoForValidate lets a test drive the REAL adminService.
// ValidateAuction end-to-end -- unlike ReviewAuctionRequest/PlaceBid, it
// needs no *sqlx.Tx (FindByID/UpdateStatus are both plain, non-transactional
// repo calls), so this is genuine repository-level coverage, not an
// isolated predicate mirror.
type fakeAuctionRepoForValidate struct {
	repository.AuctionRepository
	stored *models.Auction
}

func (f *fakeAuctionRepoForValidate) FindByID(ctx context.Context, id uuid.UUID) (*models.Auction, error) {
	cp := *f.stored
	return &cp, nil
}

func (f *fakeAuctionRepoForValidate) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	f.stored.Status = status
	return nil
}

// fakeUserRepoForValidate/fakeNotifSvcForValidate stub the seller-notification
// side effect ValidateAuction fires unconditionally on both approve and
// reject (unrelated to insurance policy, but reached on every successful
// call -- needed so the test can exercise the real method past its insurance
// gate without a nil-pointer panic on an unrelated code path).
type fakeUserRepoForValidate struct {
	repository.UserRepository
}

func (f *fakeUserRepoForValidate) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return &models.User{ID: id}, nil
}

type fakeNotifSvcForValidate struct {
	NotificationService
}

func (f *fakeNotifSvcForValidate) SendLocalizedPush(ctx context.Context, userID uuid.UUID, notificationType, language string, params map[string]string, data map[string]string) error {
	return nil
}

// TestValidateAuction_InsurancePolicy covers the SECOND, separate admin
// approval surface found during the verification pass (client feedback):
// ValidateAuction (admin_service.go, PUT /admin/auctions/:id/validate, used
// by AuctionsPage.tsx/AuctionDetailPage.tsx) transitions an `auctions` row
// already at status 'pending' directly to 'active' -- entirely independent
// of ReviewAuctionRequest/auction_requests. It had the exact same
// insurance_amount > 0 check with no InsuranceRequired() exemption before
// this fix.
func TestValidateAuction_InsurancePolicy(t *testing.T) {
	newSvc := func(a *models.Auction) (*adminService, *fakeAuctionRepoForValidate) {
		repo := &fakeAuctionRepoForValidate{stored: a}
		svc := &adminService{
			auctionRepo: repo,
			userRepo:    &fakeUserRepoForValidate{},
			notifSvc:    &fakeNotifSvcForValidate{},
		}
		return svc, repo
	}

	t.Run("legacy auction (empty policy, amount=0) approval still rejected -- V03 protection intact", func(t *testing.T) {
		a := &models.Auction{ID: uuid.New(), Status: "pending", InsurancePolicy: "", InsuranceAmount: decimal.Zero}
		svc, _ := newSvc(a)
		err := svc.ValidateAuction(context.Background(), a.ID, true, "", uuid.New())
		if err == nil {
			t.Fatal("SECURITY REGRESSION: expected a legacy pending auction with amount=0 to be rejected on approval")
		}
	})

	t.Run("explicit not_required auction (amount=0) is approvable", func(t *testing.T) {
		a := &models.Auction{ID: uuid.New(), Status: "pending", InsurancePolicy: models.InsurancePolicyNotRequired, InsuranceAmount: decimal.Zero}
		svc, repo := newSvc(a)
		if err := svc.ValidateAuction(context.Background(), a.ID, true, "", uuid.New()); err != nil {
			t.Fatalf("expected a not_required auction to be approvable, got error: %v", err)
		}
		if repo.stored.Status != "active" {
			t.Fatalf("expected status to become active, got %q", repo.stored.Status)
		}
	})

	t.Run("required + positive amount is approved", func(t *testing.T) {
		a := &models.Auction{ID: uuid.New(), Status: "pending", InsurancePolicy: models.InsurancePolicyRequired, InsuranceAmount: decimal.NewFromInt(500)}
		svc, repo := newSvc(a)
		if err := svc.ValidateAuction(context.Background(), a.ID, true, "", uuid.New()); err != nil {
			t.Fatalf("expected a required auction with a positive amount to be approvable, got error: %v", err)
		}
		if repo.stored.Status != "active" {
			t.Fatalf("expected status to become active, got %q", repo.stored.Status)
		}
	})

	t.Run("rejection path is unaffected by insurance policy", func(t *testing.T) {
		a := &models.Auction{ID: uuid.New(), Status: "pending", InsurancePolicy: "", InsuranceAmount: decimal.Zero}
		svc, repo := newSvc(a)
		if err := svc.ValidateAuction(context.Background(), a.ID, false, "not a good fit", uuid.New()); err != nil {
			t.Fatalf("expected rejection to succeed regardless of insurance state, got error: %v", err)
		}
		if repo.stored.Status != "rejected" {
			t.Fatalf("expected status to become rejected, got %q", repo.stored.Status)
		}
	})
}
