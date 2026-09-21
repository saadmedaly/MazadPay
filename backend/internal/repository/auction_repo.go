package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	apperr "github.com/mazadpay/backend/internal/errors"
	"github.com/mazadpay/backend/internal/models"
	"github.com/shopspring/decimal"
)

type AuctionFilters struct {
	Status     string
	City       string
	CategoryID int
	Query      string
	SellerID   *uuid.UUID
	WinnerID   *uuid.UUID
	UserID     *uuid.UUID
	// MarketCountryISO (migration 000046, V1 country-scoped market): when set,
	// restricts the listing to auctions whose market_country_iso equals this value
	// (legacy NULL rows are treated as DefaultAccountCountryISO, matching
	// Auction.EffectiveMarketCountryISO). Empty string = no market filtering
	// (admin/cross-market views). Never derive this from currency alone.
	MarketCountryISO string
	// Page/PerPage : pagination du listing public GET /auctions (Public Endpoints /
	// Scraping Protection — évite qu'un seul appel ne retourne la table entière).
	// 0 = valeur non fournie, la couche handler applique les valeurs par défaut/clamp.
	Page    int
	PerPage int
	// Note #1 (client feedback): the Active Auctions screen's advanced-filter
	// sheet (price range + sort mode) was already fully wired on the mobile
	// side (all_auctions_page.dart sends min_price/max_price/sort_by), but
	// GET /auctions never read or applied any of them -- confirmed by
	// inspecting AuctionHandler.List (no c.Query("min_price"/"max_price"/
	// "sort_by") at all) and this FindAll (no price WHERE clause, ORDER BY
	// hardcoded to is_featured DESC, created_at DESC). nil = not requested,
	// matching every existing filter field's own convention (0/""=unset).
	MinPrice *int
	MaxPrice *int
	// SortBy: "newest" (default, same as the pre-existing hardcoded order),
	// "price_asc", "price_desc", "ending_soon". Any other/unrecognized value
	// falls back to "newest" -- never a raw/unvalidated string reaches SQL.
	SortBy string
}

type AuctionRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*models.Auction, error)
	FindByIDTx(ctx context.Context, tx *sqlx.Tx, id uuid.UUID) (*models.Auction, error)
	// FindAll's second return value is the total count matching f (ignoring
	// pagination) -- added for client feedback #12's Active/Ended tab counts,
	// which must reflect the real total, not just the current page's length.
	FindAll(ctx context.Context, f AuctionFilters) ([]models.Auction, int, error)

	Create(ctx context.Context, tx *sqlx.Tx, a *models.Auction) error
	UpdatePrice(ctx context.Context, tx *sqlx.Tx, id uuid.UUID, newPrice decimal.Decimal, version int) (bool, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	SetWinner(ctx context.Context, tx *sqlx.Tx, id, winnerID, winningBidID uuid.UUID) error
	// TrySetWinnerAtomically/TryEndAuctionAtomically (Customer #23): the
	// race-safe canonical-finalizer primitives -- see their doc comments on
	// the auctionRepo implementation for why WHERE status = 'active' is
	// itself the concurrency guard.
	TrySetWinnerAtomically(ctx context.Context, tx *sqlx.Tx, id, winnerID, winningBidID uuid.UUID) (won bool, err error)
	TryEndAuctionAtomically(ctx context.Context, tx *sqlx.Tx, id uuid.UUID) (won bool, err error)
	// TryCancelAuctionAtomically/TryRelistAuctionAtomically/TryBuyNowAtomically
	// (Bug J): the race-safe lifecycle-transition primitives replacing the
	// generic Update for Cancel/Relist/BuyNow -- see their doc comments on
	// the auctionRepo implementation for exact semantics and guards.
	TryCancelAuctionAtomically(ctx context.Context, tx *sqlx.Tx, id uuid.UUID, reason string) (won bool, err error)
	TryRelistAuctionAtomically(ctx context.Context, tx *sqlx.Tx, id uuid.UUID, newEndTime time.Time, newPrice decimal.Decimal) (won bool, err error)
	TryBuyNowAtomically(ctx context.Context, tx *sqlx.Tx, id, buyerID uuid.UUID, buyNowPrice decimal.Decimal) (won bool, err error)
	// TryClaimBidTurn (Customer #38): the race-safe replacement for the old
	// permanent auction_bid_participants "one bid per user ever" rule --
	// atomically checks "is the CURRENT last_bidder_id different from
	// userID" and, if so, claims the turn by setting it, in one guarded
	// UPDATE. Same idiom as TrySetWinnerAtomically: the WHERE clause IS the
	// concurrency guard, never a prior read trusted across a race window.
	TryClaimBidTurn(ctx context.Context, tx *sqlx.Tx, id, userID uuid.UUID) (claimed bool, err error)
	IncrementViews(ctx context.Context, id uuid.UUID, userID *uuid.UUID) error
	IncrementBidderCount(ctx context.Context, tx *sqlx.Tx, id uuid.UUID) error
	FindExpiredActive(ctx context.Context) ([]models.Auction, error)
	GetUserHighestBid(ctx context.Context, auctionID, userID uuid.UUID) (*models.Bid, error)

	// Scheduler methods
	FindEndingBetween(ctx context.Context, start, end time.Time) ([]models.Auction, error)
	FindEndedSince(ctx context.Context, since time.Time) ([]models.Auction, error)
	GetHighestBidder(ctx context.Context, auctionID uuid.UUID) (uuid.UUID, error)

	// Admin
	ListPaginated(ctx context.Context, page, perPage int, f AuctionFilters) ([]models.Auction, int, error)
	GetStats(ctx context.Context) (int, int, int, error) // Total, Active, Pending
	ListByUserBids(ctx context.Context, userID uuid.UUID) ([]models.Auction, error)
	Update(ctx context.Context, a *models.Auction) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Images
	AddImage(ctx context.Context, img *models.AuctionImage) error
	AddImageTx(ctx context.Context, tx *sqlx.Tx, img *models.AuctionImage) error
	GetImages(ctx context.Context, auctionID uuid.UUID) ([]models.AuctionImage, error)
	DeleteImages(ctx context.Context, auctionID uuid.UUID) error
	DeleteImagesTx(ctx context.Context, tx *sqlx.Tx, auctionID uuid.UUID) error

	// Categories & Locations
	GetCategories(ctx context.Context) ([]models.Category, error)
	// GetCategoryByID (client feedback #4): single-category lookup, used to
	// stamp AuctionRequest.SubscriptionFee server-side from the category's
	// FeeTier at request-creation time.
	GetCategoryByID(ctx context.Context, id int) (*models.Category, error)
	CreateCategory(ctx context.Context, c *models.Category) error
	UpdateCategory(ctx context.Context, c *models.Category) error
	// UpdateCategoryStatus (Customer #27) is the minimal, targeted hide/show
	// write -- mirrors contentRepo.UpdateBannerStatus's exact shape (a
	// single-column UPDATE ... WHERE id, never the full-entity UpdateCategory
	// path) so toggling visibility can never accidentally overwrite
	// unrelated category fields (name/image/parent/fee_tier/etc).
	UpdateCategoryStatus(ctx context.Context, id int, isActive bool) error
	DeleteCategory(ctx context.Context, id int) error

	GetLocations(ctx context.Context) ([]models.Location, error)
	GetLocationsByCountry(ctx context.Context, countryID int) ([]models.Location, error)
	CreateLocation(ctx context.Context, l *models.Location) error
	UpdateLocation(ctx context.Context, l *models.Location) error
	DeleteLocation(ctx context.Context, id int) error

	// Countries
	GetCountries(ctx context.Context) ([]models.Country, error)
	GetCountryByCode(ctx context.Context, code string) (*models.Country, error)
	CreateCountry(ctx context.Context, c *models.Country) error
	UpdateCountry(ctx context.Context, c *models.Country) error
	DeleteCountry(ctx context.Context, id int) error
	GetCountriesWithLocations(ctx context.Context) (map[int]models.Country, error)
}

type auctionRepo struct{ db *sqlx.DB }

func NewAuctionRepository(db *sqlx.DB) AuctionRepository {
	return &auctionRepo{db: db}
}

func (r *auctionRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.Auction, error) {
	return r.findByIDInternal(ctx, r.db, id)
}

func (r *auctionRepo) FindByIDTx(ctx context.Context, tx *sqlx.Tx, id uuid.UUID) (*models.Auction, error) {
	return r.findByIDInternal(ctx, tx, id)
}

// findByIDInternal is a helper that works with both DB and Tx.
//
// Image bug fix (client feedback, GET /auctions/:id serialization audit):
// this query previously had no image_urls subquery at all, unlike FindAll/
// ListPaginated -- so a.ImageURLs was always nil, and Auction.GetImagesArray()
// (used by AuctionHandler.GetByID to build the "image_urls" field inside the
// "auction" object) always returned an empty slice regardless of how many
// rows actually existed in auction_images. Confirmed live: an auction with a
// real, persisted auction_images row showed image_urls: [] in
// GET /auctions/:id while the endpoint's separate top-level "images" array
// (built from AuctionService.GetByID's own GetImages call, unaffected by this
// bug) correctly showed the same image. Mirrors the exact subquery already
// used by FindAll/ListPaginated so both response shapes stay consistent for
// existing Flutter/web clients -- purely additive, no existing field removed
// or renamed.
func (r *auctionRepo) findByIDInternal(ctx context.Context, db sqlx.ExtContext, id uuid.UUID) (*models.Auction, error) {
	var a models.Auction
	err := db.QueryRowxContext(ctx, `
        SELECT a.*,
               c.name_ar as category_name_ar,
               l.city_name_ar as city_name_ar,
               (SELECT string_agg(ai.url, ',' ORDER BY ai.display_order)
                FROM auction_images ai WHERE ai.auction_id = a.id) as image_urls
        FROM auctions a
        LEFT JOIN categories c ON a.category_id = c.id
        LEFT JOIN locations l ON a.location_id = l.id
        WHERE a.id = $1
    `, id).StructScan(&a)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperr.ErrNotFound
		}
		return nil, err
	}
	return &a, nil
}

func (r *auctionRepo) FindAll(ctx context.Context, f AuctionFilters) ([]models.Auction, int, error) {
	where := "WHERE 1=1"
	args := []interface{}{}
	i := 1

	if f.Status != "" {
		where += fmt.Sprintf(" AND status = $%d", i)
		args = append(args, f.Status)
		i++
	}
	if f.Query != "" {
		where += fmt.Sprintf(" AND (title_ar ILIKE $%d OR title_fr ILIKE $%d OR title_en ILIKE $%d OR description_ar ILIKE $%d OR description_fr ILIKE $%d OR description_en ILIKE $%d)", i, i, i, i, i, i)
		args = append(args, "%"+f.Query+"%")
		i++
	}
	if f.CategoryID > 0 {
		where += fmt.Sprintf(" AND category_id = $%d", i)
		args = append(args, f.CategoryID)
		i++
	}
	if f.MarketCountryISO != "" {
		// Legacy rows (market_country_iso IS NULL, predating migration 000046) are
		// treated as DefaultAccountCountryISO ('MR') here, matching
		// Auction.EffectiveMarketCountryISO()'s fallback -- so a pre-migration MR
		// auction still appears in the MR market listing.
		where += fmt.Sprintf(" AND COALESCE(market_country_iso, '%s') = $%d", models.DefaultAccountCountryISO, i)
		args = append(args, f.MarketCountryISO)
		i++
	}
	// Hide expired auctions from the general/active listing only. This
	// unconditional clause predates SetWinner ever being called anywhere
	// (client feedback #10/#11 history) -- at the time, status='ended' never
	// occurred in practice, so hiding anything past end_time was harmless.
	// Now that SetWinner really persists status='ended' (with end_time
	// necessarily in the past, by definition), this clause was silently
	// excluding every single ended auction from a status='ended' request --
	// the exact query GET /auctions?status=ended relies on for the mobile
	// "Ended" filter tab (client feedback #12). Skipped only when the caller
	// explicitly asked for ended auctions; every other filter (active, or no
	// status at all) keeps the original expired-hiding behavior unchanged.
	if f.Status != "ended" {
		where += " AND end_time > NOW()"
	}
	// Note #1: advanced-filter price range, applied against current_price
	// (the live/displayed price, same column the mobile list card already
	// renders) -- not start_price, so a filter matches what the user
	// actually sees on each card.
	if f.MinPrice != nil {
		where += fmt.Sprintf(" AND current_price >= $%d", i)
		args = append(args, *f.MinPrice)
		i++
	}
	if f.MaxPrice != nil {
		where += fmt.Sprintf(" AND current_price <= $%d", i)
		args = append(args, *f.MaxPrice)
		i++
	}

	// Total matching f, ignoring pagination -- computed before LIMIT/OFFSET
	// are appended to args below (client feedback #12: real Active/Ended tab
	// counts, not currentPage.length).
	var total int
	if err := r.db.GetContext(ctx, &total, fmt.Sprintf("SELECT COUNT(*) FROM auctions a %s", where), args...); err != nil {
		return nil, 0, fmt.Errorf("failed to count auctions: %w", err)
	}

	// Pagination : page/per_page par défaut 1/25, plafonné à 100 (Public Endpoints /
	// Scraping Protection — un appel ne peut plus retourner la table entière).
	page := f.Page
	if page < 1 {
		page = 1
	}
	perPage := f.PerPage
	if perPage < 1 {
		perPage = 25
	}
	if perPage > 100 {
		perPage = 100
	}
	offset := (page - 1) * perPage
	args = append(args, perPage, offset)
	limitOffset := fmt.Sprintf(" LIMIT $%d OFFSET $%d", i, i+1)

	// Note #1: sort mode, featured-first then the caller's chosen secondary
	// order -- EXCEPT "ending_soon", whose entire point is a strict
	// nearest-expiry-first order (client requirement: an auction ending in 5
	// minutes must be shown before one ending in 30 minutes, full stop). A
	// featured-but-not-soon-to-end auction sorting above a non-featured
	// about-to-expire one would silently break that promise, so
	// "ending_soon" alone drops the is_featured DESC prefix entirely. Every
	// other sort mode (including the "newest" default) keeps the original
	// featured-first behavior unchanged. f.SortBy is matched against a fixed
	// allowlist, never interpolated directly -- an unrecognized/empty value
	// keeps the exact original "newest" order (created_at DESC), so this is
	// purely additive.
	secondaryOrder := "a.created_at DESC"
	orderBy := "a.is_featured DESC, %s"
	switch f.SortBy {
	case "price_asc":
		secondaryOrder = "a.current_price ASC"
	case "price_desc":
		secondaryOrder = "a.current_price DESC"
	case "ending_soon":
		secondaryOrder = "a.end_time ASC"
		orderBy = "%s"
	}

	rows, err := r.db.QueryxContext(ctx,
		fmt.Sprintf(`
            SELECT a.*,
                   c.name_ar as category_name_ar,
                   l.city_name_ar as city_name_ar,
                   COALESCE(
                       (SELECT string_agg(url, ',' ORDER BY display_order ASC)
                        FROM auction_images
                        WHERE auction_id = a.id),
                       ''
                   ) as image_urls
            FROM auctions a
            LEFT JOIN categories c ON a.category_id = c.id
            LEFT JOIN locations l ON a.location_id = l.id
            %s
            ORDER BY `+orderBy+`%s`, where, secondaryOrder, limitOffset),
		args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var auctions []models.Auction
	for rows.Next() {
		var a models.Auction
		if err := rows.StructScan(&a); err != nil {
			return nil, 0, fmt.Errorf("failed to scan auction: %w", err)
		}
		auctions = append(auctions, a)
	}
	return auctions, total, nil
}

func (r *auctionRepo) Create(ctx context.Context, tx *sqlx.Tx, a *models.Auction) error {
	query := `
        INSERT INTO auctions
            (id, seller_id, category_id, location_id, title_ar, title_fr, title_en, description_ar, description_fr, description_en,
             start_price, current_price, min_increment, insurance_amount, insurance_policy,
             start_time, end_time, status, lot_number, phone_contact, item_details, buy_now_price,
             market_country_iso, currency_code)
        VALUES
            (:id, :seller_id, :category_id, :location_id, :title_ar, :title_fr, :title_en, :description_ar, :description_fr, :description_en,
             :start_price, :current_price, :min_increment, :insurance_amount, :insurance_policy,
             :start_time, :end_time, :status, :lot_number, :phone_contact, :item_details, :buy_now_price,
             :market_country_iso, :currency_code)
    `
	if tx != nil {
		_, err := tx.NamedExecContext(ctx, query, a)
		return err
	}
	_, err := r.db.NamedExecContext(ctx, query, a)
	return err
}

// TryClaimBidTurn (Customer #38): atomically verifies the incoming bidder is
// NOT the current last_bidder_id (i.e. not placing two consecutive bids)
// and, if so, claims the turn in the same UPDATE. IS DISTINCT FROM handles
// the NULL case (no bids yet) correctly -- unlike `!=`, it treats NULL as
// distinct from any concrete userID, so an auction's first-ever bid is
// always claimable. Called BEFORE UpdatePrice in the same transaction, as
// its own atomic step, so UpdatePrice's own optimistic-version-conflict
// path (ErrBidConflict, a legitimate "retry") stays completely separate
// from this rule's rejection (ErrDuplicateBidder, not a retry -- the user
// must wait for someone else to bid).
func (r *auctionRepo) TryClaimBidTurn(ctx context.Context, tx *sqlx.Tx, id, userID uuid.UUID) (bool, error) {
	var returnedID uuid.UUID
	err := tx.GetContext(ctx, &returnedID, `
		UPDATE auctions SET last_bidder_id = $1
		WHERE id = $2 AND status = 'active' AND (last_bidder_id IS DISTINCT FROM $1)
		RETURNING id`,
		userID, id)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// UpdatePrice — verrouillage optimiste. Retourne false si version conflict.
func (r *auctionRepo) UpdatePrice(ctx context.Context, tx *sqlx.Tx, id uuid.UUID, newPrice decimal.Decimal, version int) (bool, error) {
	result, err := tx.ExecContext(ctx,
		`UPDATE auctions SET current_price = $1, version = version + 1, bidder_count = bidder_count + 1
         WHERE id = $2 AND version = $3 AND status = 'active'`,
		newPrice, id, version)
	if err != nil {
		return false, err
	}
	n, _ := result.RowsAffected()
	return n == 1, nil
}

func (r *auctionRepo) IncrementViews(ctx context.Context, id uuid.UUID, userID *uuid.UUID) error {
	if userID == nil {
		// Anonymous: always increment (no dedup possible)
		_, err := r.db.ExecContext(ctx, `UPDATE auctions SET views = views + 1 WHERE id = $1`, id)
		return err
	}
	// Insert into auction_views — ON CONFLICT means this user already viewed, skip
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO auction_views (auction_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		id, *userID)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 1 {
		// New unique view — increment counter
		_, err = r.db.ExecContext(ctx, `UPDATE auctions SET views = views + 1 WHERE id = $1`, id)
	}
	return err
}

func (r *auctionRepo) IncrementBidderCount(ctx context.Context, tx *sqlx.Tx, id uuid.UUID) error {
	_, err := tx.ExecContext(ctx, `UPDATE auctions SET bidder_count = bidder_count + 1 WHERE id = $1`, id)
	return err
}

func (r *auctionRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE auctions SET status = $1 WHERE id = $2`, status, id)
	return err
}

func (r *auctionRepo) SetWinner(ctx context.Context, tx *sqlx.Tx, id, winnerID, winningBidID uuid.UUID) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE auctions SET winner_id = $1, winning_bid_id = $2, status = 'ended',
         payment_deadline = now() + interval '48 hours'
         WHERE id = $3`,
		winnerID, winningBidID, id)
	return err
}

// TrySetWinnerAtomically (Customer #23 canonical finalizer) is the race-safe
// replacement for SetWinner when called from a background closer that may
// run concurrently with another closer for the same auction. The WHERE
// status = 'active' clause IS the race guard: under Postgres's default READ
// COMMITTED isolation, when two transactions concurrently run this same
// UPDATE for the same row, the second to acquire the row lock blocks until
// the first commits, then re-evaluates WHERE status = 'active' against the
// now-committed row -- which the first caller just changed to 'ended' -- so
// it matches zero rows and RETURNING yields none. Exactly one caller ever
// observes won=true for a given auction; every other concurrent/later
// caller safely no-ops. This is why correctness here does not depend on
// Redis or any application-level lock -- the guarantee comes from the
// database's own transactional row-level locking.
func (r *auctionRepo) TrySetWinnerAtomically(ctx context.Context, tx *sqlx.Tx, id, winnerID, winningBidID uuid.UUID) (won bool, err error) {
	var returnedID uuid.UUID
	err = tx.GetContext(ctx, &returnedID, `
		UPDATE auctions SET winner_id = $1, winning_bid_id = $2, status = 'ended',
         payment_deadline = now() + interval '48 hours'
         WHERE id = $3 AND status = 'active'
         RETURNING id`,
		winnerID, winningBidID, id)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// TryEndAuctionAtomically (Customer #23 canonical finalizer) is the no-bid
// counterpart of TrySetWinnerAtomically: no winner exists, so only the
// status transition itself needs the same race guard (WHERE status =
// 'active' ... RETURNING id) -- see TrySetWinnerAtomically's comment for why
// this alone is sufficient for exactly-once semantics without Redis.
func (r *auctionRepo) TryEndAuctionAtomically(ctx context.Context, tx *sqlx.Tx, id uuid.UUID) (won bool, err error) {
	var returnedID uuid.UUID
	err = tx.GetContext(ctx, &returnedID, `
		UPDATE auctions SET status = 'ended'
         WHERE id = $1 AND status = 'active'
         RETURNING id`,
		id)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// TryCancelAuctionAtomically (Bug J) is the race-safe replacement for
// CancelAuction's prior use of the generic Update, whose SQL column list
// never included status/rejection_reason -- cancellation used to silently
// no-op on the DB while returning no error and still emitting a realtime
// "canceled" event for a row that never actually changed. The WHERE clause
// is both the correctness guard (only a legal source status may transition)
// and the race guard (see TrySetWinnerAtomically's comment for the exact
// mechanism): matches the EXISTING business rule pinned by CancelAuction's
// own prior precondition (`status == "ended" || status == "canceled"` were
// the only disallowed source statuses) -- i.e. both "pending" and "active"
// auctions may be cancelled, never "ended" or already-"canceled" ones.
func (r *auctionRepo) TryCancelAuctionAtomically(ctx context.Context, tx *sqlx.Tx, id uuid.UUID, reason string) (won bool, err error) {
	var returnedID uuid.UUID
	err = tx.GetContext(ctx, &returnedID, `
		UPDATE auctions SET status = 'canceled', rejection_reason = $1
         WHERE id = $2 AND status NOT IN ('ended', 'canceled')
         RETURNING id`,
		reason, id)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// TryRelistAuctionAtomically (Bug J) is the race-safe replacement for
// RelistAuction's prior use of the generic Update. Clears ALL winner-
// finalization state from the auction's previous run (winner_id,
// winning_bid_id, payment_deadline) together -- not just winner_id -- since
// winning_bid_id/payment_deadline are themselves finalization state tied to
// that same prior winner_id; leaving either behind would attach a stale old-
// run winner reference to a freshly relisted auction. Historical bid ROWS
// are never touched here (relist has never deleted/reset bids, and Bug J
// does not change that). Guard matches RelistAuction's own existing
// precondition (`status IN ("canceled", "ended")`).
func (r *auctionRepo) TryRelistAuctionAtomically(ctx context.Context, tx *sqlx.Tx, id uuid.UUID, newEndTime time.Time, newPrice decimal.Decimal) (won bool, err error) {
	var returnedID uuid.UUID
	err = tx.GetContext(ctx, &returnedID, `
		UPDATE auctions SET status = 'pending', end_time = $1, current_price = $2,
         winner_id = NULL, winning_bid_id = NULL, payment_deadline = NULL
         WHERE id = $3 AND status IN ('canceled', 'ended')
         RETURNING id`,
		newEndTime, newPrice, id)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// TryBuyNowAtomically (Bug J) is the race-safe replacement for BuyNow's
// prior use of the generic Update. Reuses the exact WHERE status = 'active'
// ... RETURNING id race guard established by TrySetWinnerAtomically (see its
// comment for the mechanism), so BuyNow-vs-expiration and concurrent
// double-BuyNow races are both closed the same proven way. winning_bid_id is
// deliberately left NULL -- a Buy Now purchase has no real Bid row (it is an
// instant purchase, not a winning bid), matching BuyNow's own original
// intent; a fake Bid row is never invented. payment_deadline is deliberately
// NOT set here (Bug J is persistence/race-correctness only, not a new
// BuyNow business rule -- see the accompanying audit finding).
func (r *auctionRepo) TryBuyNowAtomically(ctx context.Context, tx *sqlx.Tx, id, buyerID uuid.UUID, buyNowPrice decimal.Decimal) (won bool, err error) {
	var returnedID uuid.UUID
	err = tx.GetContext(ctx, &returnedID, `
		UPDATE auctions SET winner_id = $1, status = 'ended', current_price = $2,
         winning_bid_id = NULL
         WHERE id = $3 AND status = 'active'
         RETURNING id`,
		buyerID, buyNowPrice, id)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *auctionRepo) FindExpiredActive(ctx context.Context) ([]models.Auction, error) {
	var auctions []models.Auction
	err := r.db.SelectContext(ctx, &auctions,
		`SELECT * FROM auctions WHERE status = 'active' AND end_time <= now()`)
	return auctions, err
}

func (r *auctionRepo) GetUserHighestBid(ctx context.Context, auctionID, userID uuid.UUID) (*models.Bid, error) {
	var bid models.Bid
	// bidder_name/bidder_phone are legacy denormalized columns that are
	// often NULL; models.Bid scans them as non-nullable string, so a raw
	// SELECT * fails whenever either is NULL (client feedback #19
	// hardening: this was blocking GetBidStatus's has_bid for a real bid).
	err := r.db.GetContext(ctx, &bid,
		`SELECT id, auction_id, user_id, amount, previous_price, is_winning,
                COALESCE(bidder_name, '') as bidder_name, COALESCE(bidder_phone, '') as bidder_phone,
                is_anonymous, created_at
         FROM bids WHERE auction_id = $1 AND user_id = $2 ORDER BY amount DESC LIMIT 1`,
		auctionID, userID)
	if err != nil {
		return nil, err
	}
	return &bid, nil
}

func (r *auctionRepo) AddImage(ctx context.Context, img *models.AuctionImage) error {
	return r.addImageInternal(ctx, r.db, img)
}

func (r *auctionRepo) AddImageTx(ctx context.Context, tx *sqlx.Tx, img *models.AuctionImage) error {
	return r.addImageInternal(ctx, tx, img)
}

func (r *auctionRepo) addImageInternal(ctx context.Context, db sqlx.ExtContext, img *models.AuctionImage) error {
	_, err := db.ExecContext(ctx,
		`INSERT INTO auction_images (auction_id, url, media_type, display_order)
         VALUES ($1, $2, $3, $4)`,
		img.AuctionID, img.URL, img.MediaType, img.DisplayOrder)
	return err
}

func (r *auctionRepo) GetImages(ctx context.Context, auctionID uuid.UUID) ([]models.AuctionImage, error) {
	var imgs []models.AuctionImage
	err := r.db.SelectContext(ctx, &imgs,
		`SELECT * FROM auction_images WHERE auction_id = $1 ORDER BY display_order`, auctionID)
	return imgs, err
}

func (r *auctionRepo) DeleteImages(ctx context.Context, auctionID uuid.UUID) error {
	return r.deleteImagesInternal(ctx, r.db, auctionID)
}

func (r *auctionRepo) DeleteImagesTx(ctx context.Context, tx *sqlx.Tx, auctionID uuid.UUID) error {
	return r.deleteImagesInternal(ctx, tx, auctionID)
}

func (r *auctionRepo) deleteImagesInternal(ctx context.Context, db sqlx.ExtContext, auctionID uuid.UUID) error {
	_, err := db.ExecContext(ctx, `DELETE FROM auction_images WHERE auction_id = $1`, auctionID)
	return err
}

func (r *auctionRepo) GetCategories(ctx context.Context) ([]models.Category, error) {
	var cats []models.Category
	err := r.db.SelectContext(ctx, &cats, `
		SELECT
			c.id,
			c.name_ar,
			c.name_fr,
			c.name_en,
			c.parent_id,
			c.icon_name,
			c.display_order,
			c.is_active,
			c.image_url,
			c.has_subcategories,
			c.fee_tier,
			COALESCE((
				SELECT COUNT(*)
				FROM auctions a
				WHERE a.status = 'active'
				  AND (
				      a.category_id = c.id
				      OR a.category_id IN (
				          SELECT id FROM categories sub WHERE sub.parent_id = c.id
				      )
				  )
			), 0) AS auction_count,
			(SELECT COUNT(*) FROM categories sub WHERE sub.parent_id = c.id) AS subcategories_count
		FROM categories c
		ORDER BY c.display_order`)
	return cats, err
}

// GetCategoryByID (client feedback #4): looks up a single category's
// fee_tier so the subscription fee can be stamped onto a new AuctionRequest
// server-side. Kept minimal -- callers needing counts/subcategories should
// use GetCategories instead.
func (r *auctionRepo) GetCategoryByID(ctx context.Context, id int) (*models.Category, error) {
	var cat models.Category
	err := r.db.GetContext(ctx, &cat, `
		SELECT id, name_ar, name_fr, name_en, parent_id, icon_name, display_order,
		       is_active, image_url, has_subcategories, fee_tier
		FROM categories WHERE id = $1`, id)
	if err != nil {
		return nil, err
	}
	return &cat, nil
}

func (r *auctionRepo) CreateCategory(ctx context.Context, c *models.Category) error {
	// fee_tier (client feedback #4): defaults to 'standard' when the caller
	// leaves it empty -- same fallback as UpdateCategory below, so an
	// existing/legacy client that never sends fee_tier at all still creates
	// a standard category (matching the column's own DB DEFAULT), never
	// silently rejected.
	feeTier := c.FeeTier
	if feeTier == "" {
		feeTier = models.FeeTierStandard
	}
	query := `INSERT INTO categories (name_ar, name_fr, name_en, parent_id, icon_name, display_order, image_url, is_active, fee_tier)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`
	if err := r.db.QueryRowContext(ctx, query, c.NameAr, c.NameFr, c.NameEn, c.ParentID, c.IconName, c.DisplayOrder, c.ImageURL, c.IsActive, feeTier).Scan(&c.ID); err != nil {
		return err
	}
	c.FeeTier = feeTier
	return nil
}

func (r *auctionRepo) UpdateCategory(ctx context.Context, c *models.Category) error {
	// fee_tier (client feedback #4): defaults to 'standard' when the caller
	// leaves it empty, so existing admin update call sites that don't yet
	// know about this field never accidentally clear a category's tier back
	// to empty (which the DB CHECK constraint would reject anyway).
	feeTier := c.FeeTier
	if feeTier == "" {
		feeTier = models.FeeTierStandard
	}
	query := `UPDATE categories SET name_ar = $1, name_fr = $2, name_en = $3, parent_id = $4, icon_name = $5, display_order = $6, image_url = $7, is_active = $8, fee_tier = $9
              WHERE id = $10`
	_, err := r.db.ExecContext(ctx, query, c.NameAr, c.NameFr, c.NameEn, c.ParentID, c.IconName, c.DisplayOrder, c.ImageURL, c.IsActive, feeTier, c.ID)
	return err
}

// UpdateCategoryStatus (Customer #27): mirrors contentRepo.UpdateBannerStatus
// exactly -- a single-column write, never touching name/image/parent_id/
// fee_tier/display_order. Works identically for a parent category or a
// subcategory (same table, distinguished only by parent_id being NULL or
// not), matching the client's request to reuse the same hide/show behavior
// for both.
func (r *auctionRepo) UpdateCategoryStatus(ctx context.Context, id int, isActive bool) error {
	_, err := r.db.ExecContext(ctx, `UPDATE categories SET is_active = $1 WHERE id = $2`, isActive, id)
	return err
}

func (r *auctionRepo) DeleteCategory(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM categories WHERE id = $1`, id)
	return err
}

func (r *auctionRepo) GetLocations(ctx context.Context) ([]models.Location, error) {
	var locs []models.Location
	err := r.db.SelectContext(ctx, &locs, `
		SELECT DISTINCT ON (city_name_ar) *
		FROM locations
		ORDER BY city_name_ar, area_name_ar`)
	return locs, err
}

func (r *auctionRepo) GetLocationsByCountry(ctx context.Context, countryID int) ([]models.Location, error) {
	var locs []models.Location
	err := r.db.SelectContext(ctx, &locs, `
		SELECT DISTINCT ON (city_name_ar) *
		FROM locations
		WHERE country_id = $1
		ORDER BY city_name_ar, area_name_ar`, countryID)
	return locs, err
}

func (r *auctionRepo) CreateLocation(ctx context.Context, l *models.Location) error {
	query := `INSERT INTO locations (city_name_ar, city_name_fr, area_name_ar, area_name_fr, country_id) 
              VALUES ($1, $2, $3, $4, $5) RETURNING id`
	return r.db.QueryRowContext(ctx, query, l.CityNameAr, l.CityNameFr, l.AreaNameAr, l.AreaNameFr, l.CountryID).Scan(&l.ID)
}

func (r *auctionRepo) UpdateLocation(ctx context.Context, l *models.Location) error {
	query := `UPDATE locations SET city_name_ar = $1, city_name_fr = $2, area_name_ar = $3, area_name_fr = $4, country_id = $5 WHERE id = $6`
	_, err := r.db.ExecContext(ctx, query, l.CityNameAr, l.CityNameFr, l.AreaNameAr, l.AreaNameFr, l.CountryID, l.ID)
	return err
}

func (r *auctionRepo) DeleteLocation(ctx context.Context, id int) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `UPDATE auctions SET location_id = NULL WHERE location_id = $1`, id); err != nil {
		return err
	}
	// auction_requests.location_id référence aussi locations(id) (migration 000003) —
	// oublié dans la version précédente, qui ne traitait que auctions.location_id et
	// pouvait donc échouer avec une violation FK RESTRICT si le location n'était
	// référencé que par une demande d'enchère (Countries/Locations Phase 6).
	if _, err := tx.ExecContext(ctx, `UPDATE auction_requests SET location_id = NULL WHERE location_id = $1`, id); err != nil {
		return err
	}

	result, err := tx.ExecContext(ctx, `DELETE FROM locations WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return apperr.ErrNotFound
	}

	return tx.Commit()
}

// ============================================================
// Countries Methods
// ============================================================

func (r *auctionRepo) GetCountries(ctx context.Context) ([]models.Country, error) {
	var countries []models.Country
	err := r.db.SelectContext(ctx, &countries, `
        SELECT id, code, country_code, name_ar, name_fr, name_en, flag_emoji, is_active, created_at,
               phone_min_length, phone_max_length, currency_code
        FROM countries
        WHERE is_active = TRUE
        ORDER BY name_ar ASC
    `)
	return countries, err
}

func (r *auctionRepo) GetCountryByCode(ctx context.Context, code string) (*models.Country, error) {
	var country models.Country
	err := r.db.GetContext(ctx, &country, `
        SELECT id, code, country_code, name_ar, name_fr, name_en, flag_emoji, is_active, created_at, currency_code
        FROM countries
        WHERE code = $1 AND is_active = TRUE
    `, code)
	if err != nil {
		return nil, apperr.ErrNotFound
	}
	return &country, nil
}

func (r *auctionRepo) CreateCountry(ctx context.Context, c *models.Country) error {
	query := `
        INSERT INTO countries (code, country_code, name_ar, name_fr, name_en, flag_emoji, is_active) 
        VALUES ($1, $2, $3, $4, $5, $6, $7) 
        RETURNING id
    `
	return r.db.QueryRowContext(ctx, query, c.Code, c.CountryCode, c.NameAr, c.NameFr, c.NameEn, c.FlagEmoji, c.IsActive).Scan(&c.ID)
}

func (r *auctionRepo) UpdateCountry(ctx context.Context, c *models.Country) error {
	query := `
        UPDATE countries 
        SET code = $1, country_code = $2, name_ar = $3, name_fr = $4, name_en = $5, flag_emoji = $6, is_active = $7, updated_at = CURRENT_TIMESTAMP 
        WHERE id = $8
    `
	_, err := r.db.ExecContext(ctx, query, c.Code, c.CountryCode, c.NameAr, c.NameFr, c.NameEn, c.FlagEmoji, c.IsActive, c.ID)
	return err
}

func (r *auctionRepo) DeleteCountry(ctx context.Context, id int) error {
	// Soft delete: marquer comme inactif plutôt que supprimer
	_, err := r.db.ExecContext(ctx, `UPDATE countries SET is_active = FALSE WHERE id = $1`, id)
	return err
}

func (r *auctionRepo) GetCountriesWithLocations(ctx context.Context) (map[int]models.Country, error) {

	rows, err := r.db.QueryxContext(ctx, `
        SELECT 
            c.id, c.code, c.name_ar, c.name_fr, c.name_en, c.flag_emoji, c.is_active,
            0 as locations_count
        FROM countries c
        WHERE c.is_active = TRUE
        ORDER BY c.created_at ASC
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int]models.Country)
	for rows.Next() {
		var country models.Country
		var locCount int
		if err := rows.Scan(&country.ID, &country.Code, &country.NameAr, &country.NameFr, &country.NameEn, &country.FlagEmoji, &country.IsActive, &locCount); err != nil {
			return nil, err
		}
		result[country.ID] = country
	}
	return result, nil
}

func (r *auctionRepo) ListPaginated(ctx context.Context, page, perPage int, f AuctionFilters) ([]models.Auction, int, error) {
	where := "WHERE 1=1"
	args := []interface{}{}
	i := 1

	if f.Status != "" {
		where += fmt.Sprintf(" AND status = $%d", i)
		args = append(args, f.Status)
		i++
	}
	if f.SellerID != nil {
		where += fmt.Sprintf(" AND seller_id = $%d", i)
		args = append(args, *f.SellerID)
		i++
	}
	if f.WinnerID != nil {
		where += fmt.Sprintf(" AND winner_id = $%d", i)
		args = append(args, *f.WinnerID)
		i++
	}
	if f.Query != "" {
		where += fmt.Sprintf(" AND (title_ar ILIKE $%d OR title_fr ILIKE $%d OR title_en ILIKE $%d OR description_ar ILIKE $%d OR description_fr ILIKE $%d OR description_en ILIKE $%d)", i, i, i, i, i, i)
		args = append(args, "%"+f.Query+"%")
		i++
	}

	var total int
	err := r.db.GetContext(ctx, &total, fmt.Sprintf("SELECT COUNT(*) FROM auctions %s", where), args...)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	query := fmt.Sprintf(`
        SELECT a.*,
               c.name_ar as category_name_ar,
               l.city_name_ar as city_name_ar,
               (SELECT string_agg(ai.url, ',' ORDER BY ai.display_order)
                FROM auction_images ai WHERE ai.auction_id = a.id) as image_urls
        FROM auctions a
        LEFT JOIN categories c ON a.category_id = c.id
        LEFT JOIN locations l ON a.location_id = l.id
        %s
        ORDER BY a.created_at DESC LIMIT $%d OFFSET $%d`,
		where, i, i+1)

	listArgs := append(args, perPage, offset)
	auctions := []models.Auction{}
	err = r.db.SelectContext(ctx, &auctions, query, listArgs...)
	return auctions, total, err
}

func (r *auctionRepo) GetStats(ctx context.Context) (int, int, int, error) {
	var total, active, pending int
	err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM auctions")
	if err != nil {
		return 0, 0, 0, err
	}
	err = r.db.GetContext(ctx, &active, "SELECT COUNT(*) FROM auctions WHERE status = 'active'")
	if err != nil {
		return 0, 0, 0, err
	}
	err = r.db.GetContext(ctx, &pending, "SELECT COUNT(*) FROM auctions WHERE status = 'pending'")
	return total, active, pending, err
}

func (r *auctionRepo) ListByUserBids(ctx context.Context, userID uuid.UUID) ([]models.Auction, error) {
	var auctions []models.Auction
	err := r.db.SelectContext(ctx, &auctions, `
        SELECT DISTINCT a.* FROM auctions a
        JOIN bids b ON a.id = b.auction_id
        WHERE b.user_id = $1
        ORDER BY a.created_at DESC`, userID)
	return auctions, err
}

func (r *auctionRepo) Update(ctx context.Context, a *models.Auction) error {
	query := `
		UPDATE auctions SET
			category_id = $1,
			location_id = $2,
			title_ar = $3,
			title_fr = $4,
			title_en = $5,
			description_ar = $6,
			description_fr = $7,
			description_en = $8,
			start_price = $9,
			min_increment = $10,
			insurance_amount = $11,
			start_time = $12,
			end_time = $13,
			phone_contact = $14,
			buy_now_price = $15,
			item_details = $16,
			version = version + 1
		WHERE id = $17`

	itemDetailsJSON, _ := json.Marshal(a.ItemDetails)

	_, err := r.db.ExecContext(ctx, query,
		a.CategoryID, a.LocationID, a.TitleAr, a.TitleFr, a.TitleEn,
		a.DescriptionAr, a.DescriptionFr, a.DescriptionEn,
		a.StartPrice, a.MinIncrement, a.InsuranceAmount,
		a.StartTime, a.EndTime, a.PhoneContact, a.BuyNowPrice, itemDetailsJSON,
		a.ID)
	return err
}

func (r *auctionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM auctions WHERE id = $1`, id)
	return err
}

// FindEndingBetween finds auctions ending between the given times
func (r *auctionRepo) FindEndingBetween(ctx context.Context, start, end time.Time) ([]models.Auction, error) {
	var auctions []models.Auction
	err := r.db.SelectContext(ctx, &auctions, `
		SELECT * FROM auctions 
		WHERE status = 'active' 
		AND end_time > $1 
		AND end_time <= $2
		ORDER BY end_time ASC
	`, start, end)
	return auctions, err
}

// FindEndedSince finds auctions that have ended since the given time
func (r *auctionRepo) FindEndedSince(ctx context.Context, since time.Time) ([]models.Auction, error) {
	var auctions []models.Auction
	err := r.db.SelectContext(ctx, &auctions, `
		SELECT * FROM auctions 
		WHERE status = 'active' 
		AND end_time > $1 
		AND end_time <= $2
		ORDER BY end_time DESC
	`, since, time.Now())
	return auctions, err
}

// GetHighestBidder returns the user ID of the highest bidder for an auction
func (r *auctionRepo) GetHighestBidder(ctx context.Context, auctionID uuid.UUID) (uuid.UUID, error) {
	var userID uuid.UUID
	err := r.db.GetContext(ctx, &userID, `
		SELECT user_id FROM bids 
		WHERE auction_id = $1 
		ORDER BY amount DESC 
		LIMIT 1
	`, auctionID)
	if err == sql.ErrNoRows {
		return uuid.Nil, nil
	}
	return userID, err
}
