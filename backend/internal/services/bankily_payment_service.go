package services

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/mazadpay/backend/internal/config"
	apperr "github.com/mazadpay/backend/internal/errors"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/payments/bankily"
	"github.com/mazadpay/backend/internal/repository"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// BankilyAPI is the subset of the Bankily client the service needs.
type BankilyAPI interface {
	Payment(ctx context.Context, p bankily.PaymentRequest) (*bankily.PaymentResult, error)
	CheckTransaction(ctx context.Context, operationID string) (*bankily.CheckResult, error)
}

// BankilyPayInput is one customer submission. Passcode is forwarded to
// Bankily once and then dropped -- it is never stored, logged or audited.
type BankilyPayInput struct {
	Amount         decimal.Decimal
	ClientPhone    string
	Passcode       string
	IdempotencyKey string
	Language       string // ar | fr | en (user preference)
}

type BankilyPaymentService interface {
	Enabled() bool
	MerchantCode() string
	Limits() (min, max float64, allowDecimals bool)
	Pay(ctx context.Context, userID uuid.UUID, in BankilyPayInput) (*models.BankilyPayment, error)
	// Recheck only ever calls /checkTransaction and applies its answer. It
	// can credit the wallet solely on a Bankily-confirmed TS -- it is never a
	// manual credit.
	Recheck(ctx context.Context, paymentID uuid.UUID, actor string, actorID uuid.UUID) (*models.BankilyPayment, error)
	GetForUser(ctx context.Context, userID, paymentID uuid.UUID) (*models.BankilyPayment, error)
	ListForUser(ctx context.Context, userID uuid.UUID) ([]models.BankilyPayment, error)
	ListAdmin(ctx context.Context, status string, page, perPage int) ([]models.BankilyPaymentAdminRow, int, error)
	ReconcileOnce(ctx context.Context) error
	RunReconciler(ctx context.Context)
}

type bankilyPaymentService struct {
	db         *sqlx.DB
	cfg        config.BankilyConfig
	api        BankilyAPI
	repo       repository.BankilyPaymentRepository
	walletRepo repository.WalletRepository
	userRepo   repository.UserRepository
	notifSvc   NotificationService
	auditSvc   AuditService
	logger     *zap.Logger
	now        func() time.Time
}

func NewBankilyPaymentService(db *sqlx.DB, cfg config.BankilyConfig, api BankilyAPI, repo repository.BankilyPaymentRepository, walletRepo repository.WalletRepository, userRepo repository.UserRepository, notifSvc NotificationService, auditSvc AuditService, logger *zap.Logger) BankilyPaymentService {
	return &bankilyPaymentService{db: db, cfg: cfg, api: api, repo: repo, walletRepo: walletRepo, userRepo: userRepo, notifSvc: notifSvc, auditSvc: auditSvc, logger: logger, now: time.Now}
}

// Mauritanian mobile numbers: 8 digits starting with 2, 3 or 4.
var (
	bankilyPhoneRe    = regexp.MustCompile(`^[234][0-9]{7}$`)
	bankilyPasscodeRe = regexp.MustCompile(`^[A-Za-z0-9]{1,32}$`)
	bankilyIdemKeyRe  = regexp.MustCompile(`^[A-Za-z0-9_-]{8,100}$`)
)

// Backoff for unresolved attempts: quick first checks, then slower.
var bankilyBackoff = []time.Duration{15 * time.Second, 30 * time.Second, time.Minute, 2 * time.Minute, 5 * time.Minute, 15 * time.Minute, 30 * time.Minute}

func bankilyNextDelay(attempts int) time.Duration {
	if attempts < len(bankilyBackoff) {
		return bankilyBackoff[attempts]
	}
	return time.Hour
}

func (s *bankilyPaymentService) Enabled() bool        { return s.cfg.Ready() && s.api != nil }
func (s *bankilyPaymentService) MerchantCode() string { return s.cfg.MerchantCode }
func (s *bankilyPaymentService) Limits() (float64, float64, bool) {
	return s.cfg.MinAmount, s.cfg.MaxAmount, s.cfg.AllowDecimalAmounts
}

// NormalizeBankilyPhone strips spaces and an optional +222/00222 prefix.
func NormalizeBankilyPhone(phone string) string {
	p := strings.NewReplacer(" ", "", "-", "").Replace(strings.TrimSpace(phone))
	p = strings.TrimPrefix(p, "+222")
	p = strings.TrimPrefix(p, "00222")
	return p
}

func bankilyLanguage(lang string) string {
	switch strings.ToLower(lang) {
	case "fr":
		return "FR"
	case "en":
		return "EN"
	default:
		return "AR"
	}
}

func (s *bankilyPaymentService) newOperationID() (string, error) {
	digits := s.cfg.OperationIDDigits
	if digits < 8 {
		digits = 8
	}
	var b strings.Builder
	b.WriteString(s.cfg.OperationIDPrefix)
	for i := 0; i < digits; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		b.WriteByte(byte('0' + n.Int64()))
	}
	return b.String(), nil
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23505"
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func (s *bankilyPaymentService) log(p *models.BankilyPayment) []zap.Field {
	return []zap.Field{
		zap.String("bankily_payment_id", p.ID.String()),
		zap.String("operation_id", p.OperationID),
		zap.String("user_id", p.UserID.String()),
		zap.String("client_phone", models.MaskPhone(p.ClientPhone)),
		zap.String("amount", p.Amount.String()),
	}
}

func (s *bankilyPaymentService) Pay(ctx context.Context, userID uuid.UUID, in BankilyPayInput) (*models.BankilyPayment, error) {
	if !s.Enabled() {
		return nil, apperr.ErrBankilyDisabled
	}
	// Once an attempt exists, its bookkeeping must complete even if the
	// mobile client disconnects mid-request; the HTTP client timeout still
	// bounds every Bankily call.
	ctx = context.WithoutCancel(ctx)
	if in.IdempotencyKey == "" {
		return nil, apperr.ErrBankilyIdempotencyRequired
	}
	phone := NormalizeBankilyPhone(in.ClientPhone)
	if !bankilyPhoneRe.MatchString(phone) {
		return nil, apperr.ErrBankilyInvalidPhone
	}
	passcode := strings.TrimSpace(in.Passcode)
	if !bankilyPasscodeRe.MatchString(passcode) {
		return nil, apperr.ErrBankilyInvalidPasscode
	}
	amount := in.Amount
	if amount.LessThanOrEqual(decimal.Zero) || amount.Exponent() < -2 ||
		(!s.cfg.AllowDecimalAmounts && !amount.Equal(amount.Truncate(0))) ||
		amount.LessThan(decimal.NewFromFloat(s.cfg.MinAmount)) ||
		amount.GreaterThan(decimal.NewFromFloat(s.cfg.MaxAmount)) {
		return nil, apperr.ErrBankilyInvalidAmount
	}
	if !bankilyIdemKeyRe.MatchString(in.IdempotencyKey) {
		return nil, apperr.ErrBadRequest
	}
	idemKey := &in.IdempotencyKey
	// Duplicate submission (double-tap, network retry): return the original
	// attempt, never start a second payment.
	if existing, err := s.repo.GetByIdempotencyKey(ctx, userID, in.IdempotencyKey); err == nil {
		if !existing.Amount.Equal(amount) || existing.ClientPhone != phone {
			return nil, apperr.ErrBankilyIdempotencyReused
		}
		return existing, nil
	} else if !errors.Is(err, apperr.ErrNotFound) {
		return nil, err
	}

	wallet, err := s.walletRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if wallet.IsDisabled {
		return nil, apperr.ErrWalletDisabled
	}
	currency := wallet.EffectiveCurrencyCode()
	if currency != "MRU" {
		return nil, apperr.ErrBankilyCurrency
	}

	p := &models.BankilyPayment{
		ID:             uuid.New(),
		UserID:         userID,
		IdempotencyKey: idemKey,
		Amount:         amount,
		CurrencyCode:   currency,
		ClientPhone:    phone,
		Language:       bankilyLanguage(in.Language),
		Status:         models.BankilyStatusInitiated,
	}
	for attempt := 0; ; attempt++ {
		if p.OperationID, err = s.newOperationID(); err != nil {
			return nil, err
		}
		err = s.repo.Create(ctx, p)
		if err == nil {
			break
		}
		if !isUniqueViolation(err) || attempt >= 4 {
			return nil, err
		}
		// Concurrent duplicate submission with the same key: hand back the
		// winner's attempt rather than starting another payment.
		if existing, gerr := s.repo.GetByIdempotencyKey(ctx, userID, *idemKey); gerr == nil {
			if !existing.Amount.Equal(amount) || existing.ClientPhone != phone {
				return nil, apperr.ErrBankilyIdempotencyReused
			}
			return existing, nil
		}
		// Otherwise an operation_id collision: draw a new one.
	}
	s.audit(ctx, p, "bankily_payment_initiated", "user", userID, nil)

	// SUBMITTED is persisted BEFORE the network call: from here on, a crash or
	// timeout leaves a row the reconciler resolves via checkTransaction.
	if err := s.repo.MarkSubmitted(ctx, p.ID); err != nil {
		return nil, err
	}
	s.logger.Info("bankily payment submitted", s.log(p)...)

	res, payErr := s.api.Payment(ctx, bankily.PaymentRequest{
		ClientPhone: phone,
		Passcode:    passcode,
		OperationID: p.OperationID,
		Amount:      p.Amount.StringFixed(amountScale(s.cfg.AllowDecimalAmounts)),
		Language:    p.Language,
	})
	passcode = ""
	in.Passcode = ""

	switch {
	case payErr != nil && errors.Is(payErr, bankily.ErrAuth):
		// Bankily never accepted the /payment request for processing.
		s.logger.Error("bankily authentication failed; payment not executed", append(s.log(p), zap.Error(payErr))...)
		s.fail(ctx, p, repository.BankilyOutcome{FailureReason: strPtr("provider_auth_failed")}, "system")
	case payErr != nil:
		// Timeout / network / 5xx / malformed: outcome UNKNOWN. Never resend
		// the payment -- verify through checkTransaction instead.
		s.logger.Warn("bankily payment outcome uncertain; verifying via checkTransaction", append(s.log(p), zap.Error(payErr))...)
		s.pending(ctx, p, repository.BankilyOutcome{FailureReason: strPtr(uncertainReason(payErr))}, 0)
		s.verify(ctx, p.ID, "system", uuid.Nil)
	default:
		s.handlePaymentResult(ctx, p, res)
	}
	return s.repo.GetByID(ctx, p.ID)
}

func amountScale(allowDecimals bool) int32 {
	if allowDecimals {
		return 2
	}
	return 0
}

func uncertainReason(err error) string {
	var ue *bankily.UncertainError
	if errors.As(err, &ue) {
		return "payment_" + ue.Reason
	}
	return "payment_uncertain"
}

// handlePaymentResult applies a definite /payment answer. Even errorCode 0 is
// confirmed through checkTransaction (TS) before any wallet credit.
func (s *bankilyPaymentService) handlePaymentResult(ctx context.Context, p *models.BankilyPayment, res *bankily.PaymentResult) {
	o := repository.BankilyOutcome{
		BankilyErrorCode:    strPtr(res.ErrorCode),
		BankilyErrorMessage: strPtr(res.ErrorMessage),
		BankilyTxID:         strPtr(res.TransactionID),
	}
	fields := append(s.log(p), zap.String("bankily_error_code", res.ErrorCode))
	switch res.ErrorCode {
	case bankily.PaymentCodeOperationIDRequired:
		s.logger.Error("bankily rejected payment: operationId required", fields...)
		o.FailureReason = strPtr("operation_id_required")
		o.NeedsReview = true
		s.fail(ctx, p, o, "system")
	case bankily.PaymentCodeOperationIDExists:
		// Undocumented by the v1.0 spec. Bankily already holds an operation
		// with this ID, so a later TS for it may not be this customer's
		// payment: never resolve it automatically.
		s.logger.Error("bankily payment: operationId already exists; attempt sent to manual review", fields...)
		o.FailureReason = strPtr("operation_id_already_exists")
		s.review(ctx, p, o)
	case bankily.PaymentCodeSuccess, bankily.PaymentCodeOtherError:
		s.logger.Info("bankily payment answered; confirming via checkTransaction", fields...)
		s.pending(ctx, p, o, 0)
		s.verifyAfterPayment(ctx, p, res)
	default:
		s.logger.Warn("bankily payment returned an undocumented errorCode; verifying", fields...)
		o.FailureReason = strPtr("payment_unknown_code")
		s.pending(ctx, p, o, 0)
		s.verify(ctx, p.ID, "system", uuid.Nil)
	}
}

// verifyAfterPayment confirms a definite /payment answer. errorCode 1 with
// no Bankily record ("not found") is a definite rejection -> FAILED; every
// other case follows checkTransaction.
func (s *bankilyPaymentService) verifyAfterPayment(ctx context.Context, p *models.BankilyPayment, res *bankily.PaymentResult) {
	check, err := s.api.CheckTransaction(ctx, p.OperationID)
	if err != nil {
		s.logger.Warn("bankily checkTransaction uncertain after payment; left for reconciler", append(s.log(p), zap.Error(err))...)
		s.countCheck(ctx, p, repository.BankilyOutcome{})
		return
	}
	if res.ErrorCode == bankily.PaymentCodeOtherError && check.Status == "" {
		s.fail(ctx, p, repository.BankilyOutcome{
			BankilyErrorCode:    strPtr(res.ErrorCode),
			BankilyErrorMessage: strPtr(res.ErrorMessage),
			FailureReason:       strPtr("payment_rejected"),
			CountCheck:          true,
		}, "system")
		return
	}
	s.applyCheck(ctx, p, check, "system", uuid.Nil)
}

// verify runs one checkTransaction for an unresolved attempt.
func (s *bankilyPaymentService) verify(ctx context.Context, id uuid.UUID, actor string, actorID uuid.UUID) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil || p.IsTerminal() {
		return
	}
	check, err := s.api.CheckTransaction(ctx, p.OperationID)
	if err != nil {
		s.logger.Warn("bankily checkTransaction uncertain; will retry", append(s.log(p), zap.Error(err))...)
		s.countCheck(ctx, p, repository.BankilyOutcome{})
		return
	}
	s.applyCheck(ctx, p, check, actor, actorID)
}

// applyCheck maps a checkTransaction answer onto the state machine:
//
//	TS -> SUCCESS (credit exactly once)   TF -> FAILED   TA -> PENDING
//	errorCode 1 / unknown status -> stays PENDING_VERIFICATION (money may have
//	moved; never auto-failed -- flagged for review after MaxPendingAge).
func (s *bankilyPaymentService) applyCheck(ctx context.Context, p *models.BankilyPayment, c *bankily.CheckResult, actor string, actorID uuid.UUID) {
	o := repository.BankilyOutcome{
		BankilyStatus: strPtr(c.Status),
		BankilyTxID:   strPtr(c.TransactionID),
		CountCheck:    true,
	}
	fields := append(s.log(p), zap.String("check_error_code", c.ErrorCode), zap.String("check_status", c.Status))
	switch {
	case c.Status == bankily.StatusSuccess:
		s.logger.Info("bankily confirmed TS; crediting wallet", fields...)
		if err := s.credit(ctx, p.ID, strPtr(c.TransactionID), strPtr(c.Status), actor, actorID); err != nil {
			s.logger.Error("bankily credit failed; will retry", append(fields, zap.Error(err))...)
			s.countCheck(ctx, p, o)
		}
	case c.Status == bankily.StatusFailed:
		s.logger.Info("bankily confirmed TF; marking failed", fields...)
		o.FailureReason = strPtr("bankily_failed")
		o.BankilyErrorMessage = strPtr(c.ErrorMessage)
		s.fail(ctx, p, o, actor)
	case c.Status == bankily.StatusPending:
		s.logger.Info("bankily reports TA (pending); not crediting", fields...)
		s.countCheck(ctx, p, o)
	default:
		s.logger.Warn("bankily checkTransaction gave no final status; still pending", fields...)
		o.BankilyErrorMessage = strPtr(c.ErrorMessage)
		s.countCheck(ctx, p, o)
	}
}

func (s *bankilyPaymentService) pending(ctx context.Context, p *models.BankilyPayment, o repository.BankilyOutcome, attempts int) {
	o.Status = models.BankilyStatusPendingVerification
	next := s.now().Add(bankilyNextDelay(attempts))
	o.NextCheckAt = &next
	if _, err := s.repo.ApplyOutcome(ctx, p.ID, o); err != nil {
		s.logger.Error("bankily: failed to persist pending state", append(s.log(p), zap.Error(err))...)
	}
}

func (s *bankilyPaymentService) countCheck(ctx context.Context, p *models.BankilyPayment, o repository.BankilyOutcome) {
	o.CountCheck = true
	s.pending(ctx, p, o, p.CheckAttempts+1)
}

// review moves an unresolved attempt to NEEDS_REVIEW: no automatic path
// (inline, reconciler, user or admin re-check) ever credits from that state.
func (s *bankilyPaymentService) review(ctx context.Context, p *models.BankilyPayment, o repository.BankilyOutcome) {
	o.Status = models.BankilyStatusNeedsReview
	o.NeedsReview = true
	changed, err := s.repo.ApplyOutcome(ctx, p.ID, o)
	if err != nil {
		s.logger.Error("bankily: failed to persist NEEDS_REVIEW state", append(s.log(p), zap.Error(err))...)
		return
	}
	if changed {
		reason := ""
		if o.FailureReason != nil {
			reason = *o.FailureReason
		}
		s.audit(ctx, p, "bankily_payment_needs_review", "system", uuid.Nil, models.JSONB{"failure_reason": reason})
	}
}

func (s *bankilyPaymentService) fail(ctx context.Context, p *models.BankilyPayment, o repository.BankilyOutcome, actor string) {
	o.Status = models.BankilyStatusFailed
	changed, err := s.repo.ApplyOutcome(ctx, p.ID, o)
	if err != nil {
		s.logger.Error("bankily: failed to persist FAILED state", append(s.log(p), zap.Error(err))...)
		return
	}
	if changed {
		reason := ""
		if o.FailureReason != nil {
			reason = *o.FailureReason
		}
		s.audit(ctx, p, "bankily_payment_failed", actor, uuid.Nil, models.JSONB{"failure_reason": reason})
	}
}

// credit is the ONLY place a Bankily payment changes a wallet balance. One
// DB transaction: lock the attempt row, re-check it is still unresolved,
// write the completed ledger row (reference 'bankily:<operation_id>', unique
// for this gateway), credit the balance, mark SUCCESS. Concurrent callers
// (inline response, reconciler, user re-check, admin re-check) serialize on
// the row lock; the loser sees SUCCESS and returns without crediting. The
// unique indexes are a second, independent barrier.
func (s *bankilyPaymentService) credit(ctx context.Context, id uuid.UUID, bankilyTxID, bankilyStatus *string, actor string, actorID uuid.UUID) error {
	dbtx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer dbtx.Rollback()

	p, err := s.repo.LockForUpdate(ctx, dbtx, id)
	if err != nil {
		return err
	}
	if p.Status == models.BankilyStatusSuccess {
		return nil // already credited by another path
	}
	if p.Status == models.BankilyStatusFailed || p.Status == models.BankilyStatusNeedsReview {
		// Bankily says TS for an attempt that is closed or under manual
		// review. Never auto-credit it: flag it for an operator.
		s.logger.Error("bankily TS received for an attempt not eligible for automatic credit; flagged for review", append(s.log(p), zap.String("local_status", p.Status))...)
		_, _ = dbtx.ExecContext(ctx, `UPDATE bankily_payments SET needs_review = TRUE, bankily_status = $2, updated_at = NOW() WHERE id = $1`, id, bankilyStatus)
		if err := dbtx.Commit(); err != nil {
			return err
		}
		s.audit(ctx, p, "bankily_payment_status_conflict", actor, actorID, models.JSONB{"bankily_status": "TS", "local_status": p.Status})
		return nil
	}

	if _, err := s.walletRepo.FindForUpdate(ctx, dbtx, p.UserID); err != nil {
		return fmt.Errorf("lock wallet: %w", err)
	}
	ledgerID := uuid.New()
	reference := "bankily:" + p.OperationID
	description := "Bankily B-PAY " + p.OperationID
	if _, err := dbtx.ExecContext(ctx, `
		INSERT INTO transactions
			(id, user_id, type, amount, gateway, status, reference, payment_method,
			 description, currency_code, net_amount)
		VALUES ($1, $2, 'deposit', $3, $4, 'completed', $5, $4, $6, $7, $3)`,
		ledgerID, p.UserID, p.Amount, models.BankilyGateway, reference, description, p.CurrencyCode); err != nil {
		if isUniqueViolation(err) {
			return apperr.ErrConflict
		}
		return err
	}
	if err := s.walletRepo.CreditBalance(ctx, dbtx, p.UserID, p.Amount); err != nil {
		return err
	}
	if err := s.repo.MarkSuccessTx(ctx, dbtx, p.ID, ledgerID, bankilyTxID, bankilyStatus); err != nil {
		return err
	}
	if err := dbtx.Commit(); err != nil {
		return err
	}

	s.logger.Info("bankily payment credited", append(s.log(p), zap.String("wallet_transaction_id", ledgerID.String()))...)
	s.audit(ctx, p, "bankily_payment_credited", actor, actorID, models.JSONB{"wallet_transaction_id": ledgerID.String()})
	s.notifyCredited(ctx, p, ledgerID)
	return nil
}

func (s *bankilyPaymentService) notifyCredited(ctx context.Context, p *models.BankilyPayment, ledgerID uuid.UUID) {
	if s.notifSvc == nil || s.userRepo == nil {
		return
	}
	language := "ar"
	if user, err := s.userRepo.FindByID(ctx, p.UserID); err == nil && user.LanguagePref != "" {
		language = user.LanguagePref
	}
	_ = s.notifSvc.SendLocalizedPush(ctx, p.UserID, "deposit_confirmed", language,
		map[string]string{"amount": p.Amount.String(), "currency": p.CurrencyCode},
		map[string]string{"type": "deposit_confirmed", "transaction_id": ledgerID.String()})
}

// audit writes a financial audit entry. Never includes the passcode; the
// phone is masked.
func (s *bankilyPaymentService) audit(ctx context.Context, p *models.BankilyPayment, action, actor string, actorID uuid.UUID, extra models.JSONB) {
	if s.auditSvc == nil {
		return
	}
	details := models.JSONB{
		"bankily_payment_id": p.ID.String(),
		"operation_id":       p.OperationID,
		"user_id":            p.UserID.String(),
		"amount":             p.Amount.String(),
		"currency":           p.CurrencyCode,
		"client_phone":       models.MaskPhone(p.ClientPhone),
	}
	for k, v := range extra {
		details[k] = v
	}
	opts := []LogOption{WithDetailsJSON(details), WithEntityKey(p.OperationID), WithActorType(actor)}
	if actorID == uuid.Nil {
		opts = append(opts, WithSystemActor())
	}
	id := p.ID
	if err := s.auditSvc.Log(ctx, actorID, action, "bankily_payment", &id,
		fmt.Sprintf("operation_id=%s amount=%s", p.OperationID, p.Amount.String()), opts...); err != nil {
		s.logger.Warn("bankily audit log failed", append(s.log(p), zap.Error(err))...)
	}
}

func (s *bankilyPaymentService) Recheck(ctx context.Context, paymentID uuid.UUID, actor string, actorID uuid.UUID) (*models.BankilyPayment, error) {
	if !s.Enabled() {
		return nil, apperr.ErrBankilyDisabled
	}
	p, err := s.repo.GetByID(ctx, paymentID)
	if err != nil {
		return nil, err
	}
	if p.Status == models.BankilyStatusNeedsReview {
		// Informational only: record what Bankily reports for the operator.
		// Nothing here can change the status or credit the wallet.
		s.audit(ctx, p, "bankily_payment_review_check", actor, actorID, nil)
		if check, err := s.api.CheckTransaction(ctx, p.OperationID); err == nil {
			if err := s.repo.RecordReviewCheck(ctx, p.ID, strPtr(check.Status)); err != nil {
				s.logger.Error("bankily: failed to record review check", append(s.log(p), zap.Error(err))...)
			}
		} else {
			s.logger.Warn("bankily review check uncertain", append(s.log(p), zap.Error(err))...)
		}
	} else if !p.IsTerminal() && p.Status != models.BankilyStatusInitiated {
		s.audit(ctx, p, "bankily_payment_recheck", actor, actorID, nil)
		s.verify(ctx, p.ID, actor, actorID)
	}
	return s.repo.GetByID(ctx, paymentID)
}

func (s *bankilyPaymentService) GetForUser(ctx context.Context, userID, paymentID uuid.UUID) (*models.BankilyPayment, error) {
	return s.repo.GetByIDForUser(ctx, paymentID, userID)
}

func (s *bankilyPaymentService) ListForUser(ctx context.Context, userID uuid.UUID) ([]models.BankilyPayment, error) {
	return s.repo.ListByUser(ctx, userID, 50)
}

func (s *bankilyPaymentService) ListAdmin(ctx context.Context, status string, page, perPage int) ([]models.BankilyPaymentAdminRow, int, error) {
	return s.repo.ListAdmin(ctx, status, page, perPage)
}

// ReconcileOnce resolves due unresolved attempts through checkTransaction.
func (s *bankilyPaymentService) ReconcileOnce(ctx context.Context) error {
	if !s.Enabled() {
		return nil
	}
	if n, err := s.repo.FailOrphanInitiated(ctx, s.now().Add(-5*time.Minute)); err != nil {
		s.logger.Error("bankily reconciler: orphan cleanup failed", zap.Error(err))
	} else if n > 0 {
		s.logger.Warn("bankily reconciler: failed never-submitted attempts", zap.Int64("count", n))
	}
	if s.cfg.MaxPendingAge > 0 {
		if n, err := s.repo.FlagStale(ctx, s.now().Add(-s.cfg.MaxPendingAge)); err == nil && n > 0 {
			s.logger.Error("bankily reconciler: attempts unresolved past max pending age, flagged for review", zap.Int64("count", n))
		}
	}
	due, err := s.repo.ClaimDue(ctx, 50, 2*time.Minute)
	if err != nil {
		return err
	}
	for i := range due {
		p := due[i]
		check, err := s.api.CheckTransaction(ctx, p.OperationID)
		if err != nil {
			s.logger.Warn("bankily reconciler: checkTransaction uncertain", append(s.log(&p), zap.Error(err))...)
			s.countCheck(ctx, &p, repository.BankilyOutcome{})
			continue
		}
		s.applyCheck(ctx, &p, check, "system", uuid.Nil)
	}
	return nil
}

func (s *bankilyPaymentService) RunReconciler(ctx context.Context) {
	interval := s.cfg.ReconcileInterval
	if interval <= 0 {
		interval = 30 * time.Second
	}
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			if err := s.ReconcileOnce(ctx); err != nil {
				s.logger.Error("bankily reconciler run failed", zap.Error(err))
			}
		}
	}
}
