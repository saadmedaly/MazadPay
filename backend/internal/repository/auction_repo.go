package repository

import (
    "context"
    "fmt"

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
    Page       int
    PerPage    int
}

type AuctionRepository interface {
    FindByID(ctx context.Context, id uuid.UUID) (*models.Auction, error)
    FindAll(ctx context.Context, f AuctionFilters) ([]models.Auction, int, error)
    Create(ctx context.Context, tx *sqlx.Tx, a *models.Auction) error
    UpdatePrice(ctx context.Context, tx *sqlx.Tx, id uuid.UUID, newPrice decimal.Decimal, version int) (bool, error)
    UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
    SetWinner(ctx context.Context, tx *sqlx.Tx, id, winnerID, winningBidID uuid.UUID) error
    IncrementViews(ctx context.Context, id uuid.UUID) error
    IncrementBidderCount(ctx context.Context, tx *sqlx.Tx, id uuid.UUID) error
    FindExpiredActive(ctx context.Context) ([]models.Auction, error)

    // Images
    AddImage(ctx context.Context, img *models.AuctionImage) error
    GetImages(ctx context.Context, auctionID uuid.UUID) ([]models.AuctionImage, error)

    // Categories & Locations
    GetCategories(ctx context.Context) ([]models.Category, error)
    GetLocations(ctx context.Context) ([]models.Location, error)
}

type auctionRepo struct{ db *sqlx.DB }

func NewAuctionRepository(db *sqlx.DB) AuctionRepository {
    return &auctionRepo{db: db}
}

func (r *auctionRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.Auction, error) {
    var a models.Auction
    err := r.db.GetContext(ctx, &a, `SELECT * FROM auctions WHERE id = $1`, id)
    if err != nil {
        return nil, apperr.ErrNotFound
    }
    return &a, nil
}

func (r *auctionRepo) FindAll(ctx context.Context, f AuctionFilters) ([]models.Auction, int, error) {
    if f.Page < 1 { f.Page = 1 }
    if f.PerPage < 1 || f.PerPage > 50 { f.PerPage = 20 }
    offset := (f.Page - 1) * f.PerPage

    where := "WHERE 1=1"
    args := []interface{}{}
    i := 1

    if f.Status != "" {
        where += fmt.Sprintf(" AND status = $%d", i)
        args = append(args, f.Status)
        i++
    }
    if f.Query != "" {
        where += fmt.Sprintf(" AND (title ILIKE $%d OR description ILIKE $%d)", i, i+1)
        args = append(args, "%"+f.Query+"%", "%"+f.Query+"%")
        i += 2
    }
    if f.CategoryID > 0 {
        where += fmt.Sprintf(" AND category_id = $%d", i)
        args = append(args, f.CategoryID)
        i++
    }

    var total int
    _ = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM auctions "+where, args...).Scan(&total)

    args = append(args, f.PerPage, offset)
    rows, err := r.db.QueryxContext(ctx,
        fmt.Sprintf("SELECT * FROM auctions %s ORDER BY is_featured DESC, created_at DESC LIMIT $%d OFFSET $%d", where, i, i+1),
        args...)
    if err != nil {
        return nil, 0, err
    }
    defer rows.Close()

    var auctions []models.Auction
    for rows.Next() {
        var a models.Auction
        if err := rows.StructScan(&a); err == nil {
            auctions = append(auctions, a)
        }
    }
    return auctions, total, nil
}

func (r *auctionRepo) Create(ctx context.Context, tx *sqlx.Tx, a *models.Auction) error {
    query := `
        INSERT INTO auctions
            (id, seller_id, category_id, location_id, title, description,
             start_price, current_price, min_increment, insurance_amount,
             end_time, status, lot_number, phone_contact, item_details, buy_now_price)
        VALUES
            (:id, :seller_id, :category_id, :location_id, :title, :description,
             :start_price, :current_price, :min_increment, :insurance_amount,
             :end_time, :status, :lot_number, :phone_contact, :item_details, :buy_now_price)
    `
    if tx != nil {
        _, err := tx.NamedExecContext(ctx, query, a)
        return err
    }
    _, err := r.db.NamedExecContext(ctx, query, a)
    return err
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

func (r *auctionRepo) IncrementViews(ctx context.Context, id uuid.UUID) error {
    _, err := r.db.ExecContext(ctx, `UPDATE auctions SET views = views + 1 WHERE id = $1`, id)
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

func (r *auctionRepo) FindExpiredActive(ctx context.Context) ([]models.Auction, error) {
    var auctions []models.Auction
    err := r.db.SelectContext(ctx, &auctions,
        `SELECT * FROM auctions WHERE status = 'active' AND end_time <= now()`)
    return auctions, err
}

func (r *auctionRepo) AddImage(ctx context.Context, img *models.AuctionImage) error {
    _, err := r.db.ExecContext(ctx,
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

func (r *auctionRepo) GetCategories(ctx context.Context) ([]models.Category, error) {
    var cats []models.Category
    err := r.db.SelectContext(ctx, &cats, `SELECT * FROM categories ORDER BY display_order`)
    return cats, err
}

func (r *auctionRepo) GetLocations(ctx context.Context) ([]models.Location, error) {
    var locs []models.Location
    err := r.db.SelectContext(ctx, &locs, `SELECT * FROM locations ORDER BY city_name, area_name`)
    return locs, err
}
