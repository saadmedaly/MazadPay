package repository

import (
	"context"
	"encoding/json"
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
	SellerID   *uuid.UUID
	WinnerID   *uuid.UUID
}

type AuctionRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*models.Auction, error)
	FindAll(ctx context.Context, f AuctionFilters) ([]models.Auction, error)

	Create(ctx context.Context, tx *sqlx.Tx, a *models.Auction) error
	UpdatePrice(ctx context.Context, tx *sqlx.Tx, id uuid.UUID, newPrice decimal.Decimal, version int) (bool, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	SetWinner(ctx context.Context, tx *sqlx.Tx, id, winnerID, winningBidID uuid.UUID) error
	IncrementViews(ctx context.Context, id uuid.UUID) error
	IncrementBidderCount(ctx context.Context, tx *sqlx.Tx, id uuid.UUID) error
	FindExpiredActive(ctx context.Context) ([]models.Auction, error)

	// Admin
	ListPaginated(ctx context.Context, page, perPage int, f AuctionFilters) ([]models.Auction, int, error)
	GetStats(ctx context.Context) (int, int, int, error) // Total, Active, Pending
	ListByUserBids(ctx context.Context, userID uuid.UUID) ([]models.Auction, error)
	Update(ctx context.Context, a *models.Auction) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Images
	AddImage(ctx context.Context, img *models.AuctionImage) error
	GetImages(ctx context.Context, auctionID uuid.UUID) ([]models.AuctionImage, error)
	DeleteImages(ctx context.Context, auctionID uuid.UUID) error

	// Categories & Locations
	GetCategories(ctx context.Context) ([]models.Category, error)
	CreateCategory(ctx context.Context, c *models.Category) error
	UpdateCategory(ctx context.Context, c *models.Category) error
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
	var a models.Auction
	err := r.db.GetContext(ctx, &a, `
        SELECT a.*, 
               c.name_ar as category_name_ar,
               l.city_name_ar as city_name_ar
        FROM auctions a
        LEFT JOIN categories c ON a.category_id = c.id
        LEFT JOIN locations l ON a.location_id = l.id
        WHERE a.id = $1
    `, id)
	if err != nil {
		return nil, apperr.ErrNotFound
	}
	return &a, nil
}

func (r *auctionRepo) FindAll(ctx context.Context, f AuctionFilters) ([]models.Auction, error) {
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

	rows, err := r.db.QueryxContext(ctx,
		fmt.Sprintf(`
            SELECT a.*, 
                   c.name_ar as category_name_ar,
                   l.city_name_ar as city_name_ar
            FROM auctions a
            LEFT JOIN categories c ON a.category_id = c.id
            LEFT JOIN locations l ON a.location_id = l.id
            %s 
            ORDER BY a.is_featured DESC, a.created_at DESC`, where),
		args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var auctions []models.Auction
	for rows.Next() {
		var a models.Auction
		if err := rows.StructScan(&a); err == nil {
			auctions = append(auctions, a)
		}
	}
	return auctions, nil
}

func (r *auctionRepo) Create(ctx context.Context, tx *sqlx.Tx, a *models.Auction) error {
	query := `
        INSERT INTO auctions
            (id, seller_id, category_id, location_id, title_ar, title_fr, title_en, description_ar, description_fr, description_en,
             start_price, current_price, min_increment, insurance_amount,
             start_time, end_time, status, lot_number, phone_contact, item_details, buy_now_price)
        VALUES
            (:id, :seller_id, :category_id, :location_id, :title_ar, :title_fr, :title_en, :description_ar, :description_fr, :description_en,
             :start_price, :current_price, :min_increment, :insurance_amount,
             :start_time, :end_time, :status, :lot_number, :phone_contact, :item_details, :buy_now_price)
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

func (r *auctionRepo) DeleteImages(ctx context.Context, auctionID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM auction_images WHERE auction_id = $1`, auctionID)
	return err
}

func (r *auctionRepo) GetCategories(ctx context.Context) ([]models.Category, error) {
	var cats []models.Category
	err := r.db.SelectContext(ctx, &cats, `SELECT * FROM categories ORDER BY display_order`)
	return cats, err
}

func (r *auctionRepo) CreateCategory(ctx context.Context, c *models.Category) error {
	query := `INSERT INTO categories (name_ar, name_fr, name_en, parent_id, icon_name, display_order)
              VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	return r.db.QueryRowContext(ctx, query, c.NameAr, c.NameFr, c.NameEn, c.ParentID, c.IconName, c.DisplayOrder).Scan(&c.ID)
}

func (r *auctionRepo) UpdateCategory(ctx context.Context, c *models.Category) error {
	query := `UPDATE categories SET name_ar = $1, name_fr = $2, name_en = $3, parent_id = $4, icon_name = $5, display_order = $6
              WHERE id = $7`
	_, err := r.db.ExecContext(ctx, query, c.NameAr, c.NameFr, c.NameEn, c.ParentID, c.IconName, c.DisplayOrder, c.ID)
	return err
}

func (r *auctionRepo) DeleteCategory(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM categories WHERE id = $1`, id)
	return err
}

func (r *auctionRepo) GetLocations(ctx context.Context) ([]models.Location, error) {
	var locs []models.Location
	err := r.db.SelectContext(ctx, &locs, `SELECT * FROM locations ORDER BY city_name_ar, area_name_ar`)
	return locs, err
}

func (r *auctionRepo) GetLocationsByCountry(ctx context.Context, countryID int) ([]models.Location, error) {
	// Pour l'instant, comme country_id n'existe pas dans la table locations,
	// nous retournons toutes les locations triées par ville
	var locs []models.Location
	err := r.db.SelectContext(ctx, &locs, `SELECT * FROM locations ORDER BY city_name_ar, area_name_ar`)
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
	_, err := r.db.ExecContext(ctx, `UPDATE auctions SET location_id = NULL WHERE location_id = $1`, id)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, `DELETE FROM locations WHERE id = $1`, id)
	return err
}

// ============================================================
// Countries Methods
// ============================================================

func (r *auctionRepo) GetCountries(ctx context.Context) ([]models.Country, error) {
	var countries []models.Country
	err := r.db.SelectContext(ctx, &countries, `
        SELECT id, code, name_ar, name_fr, name_en, flag_emoji, is_active 
        FROM countries 
        WHERE is_active = TRUE 
        ORDER BY created_at ASC
    `)
	return countries, err
}

func (r *auctionRepo) GetCountryByCode(ctx context.Context, code string) (*models.Country, error) {
	var country models.Country
	err := r.db.GetContext(ctx, &country, `
        SELECT id, code, name_ar, name_fr, name_en, flag_emoji, is_active 
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
        INSERT INTO countries (code, name_ar, name_fr, name_en, flag_emoji, is_active) 
        VALUES ($1, $2, $3, $4, $5, $6) 
        RETURNING id
    `
	return r.db.QueryRowContext(ctx, query, c.Code, c.NameAr, c.NameFr, c.NameEn, c.FlagEmoji, c.IsActive).Scan(&c.ID)
}

func (r *auctionRepo) UpdateCountry(ctx context.Context, c *models.Country) error {
	query := `
        UPDATE countries 
        SET code = $1, name_ar = $2, name_fr = $3, name_en = $4, flag_emoji = $5, is_active = $6, updated_at = CURRENT_TIMESTAMP 
        WHERE id = $7
    `
	_, err := r.db.ExecContext(ctx, query, c.Code, c.NameAr, c.NameFr, c.NameEn, c.FlagEmoji, c.IsActive, c.ID)
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
               l.city_name_ar as city_name_ar
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
