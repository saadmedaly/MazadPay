package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/mazadpay/backend/internal/models"
)

var ErrNotFound = errors.New("resource not found")

type RequestRepository interface {
	// Auction Requests
	CreateAuctionRequest(ctx context.Context, req *models.AuctionRequest) error
	GetAuctionRequests(ctx context.Context, status string, userID *uuid.UUID, dateFrom, dateTo *time.Time, categoryID, locationID *int, minPrice, maxPrice *float64, sortBy, sortOrder string, page, perPage int) ([]models.AuctionRequest, int, error)
	GetAuctionRequestByID(ctx context.Context, id uuid.UUID) (*models.AuctionRequest, error)
	GetUserAuctionRequests(ctx context.Context, userID uuid.UUID, status string, page, perPage int) ([]models.AuctionRequest, int, error)
	UpdateAuctionRequestStatus(ctx context.Context, id uuid.UUID, status, notes string, reviewedBy uuid.UUID) error
	UpdateAuctionRequestStatusTx(ctx context.Context, tx *sqlx.Tx, id uuid.UUID, status, notes string, reviewedBy uuid.UUID) error
	DeleteAuctionRequest(ctx context.Context, id uuid.UUID) error
	CountPendingAuctionRequests(ctx context.Context) (int, error)
	BulkUpdateAuctionRequestStatus(ctx context.Context, ids []uuid.UUID, status, notes string, reviewedBy uuid.UUID) error
	BulkDeleteAuctionRequests(ctx context.Context, ids []uuid.UUID) error

	// Banner Requests
	CreateBannerRequest(ctx context.Context, req *models.BannerRequest) error
	GetBannerRequests(ctx context.Context, status string, userID *uuid.UUID, dateFrom, dateTo *time.Time, sortBy, sortOrder string, page, perPage int) ([]models.BannerRequest, int, error)
	GetBannerRequestByID(ctx context.Context, id uuid.UUID) (*models.BannerRequest, error)
	GetUserBannerRequests(ctx context.Context, userID uuid.UUID, status string, page, perPage int) ([]models.BannerRequest, int, error)
	UpdateBannerRequestStatus(ctx context.Context, id uuid.UUID, status, notes string, reviewedBy uuid.UUID) error
	UpdateBannerRequestStatusTx(ctx context.Context, tx *sqlx.Tx, id uuid.UUID, status, notes string, reviewedBy uuid.UUID) error
	DeleteBannerRequest(ctx context.Context, id uuid.UUID) error
	CountPendingBannerRequests(ctx context.Context) (int, error)
	BulkUpdateBannerRequestStatus(ctx context.Context, ids []uuid.UUID, status, notes string, reviewedBy uuid.UUID) error
	BulkDeleteBannerRequests(ctx context.Context, ids []uuid.UUID) error

	// Transaction support
	BeginTx(ctx context.Context) (*sqlx.Tx, error)
}

type requestRepo struct {
	db *sqlx.DB
}

func NewRequestRepository(db *sqlx.DB) RequestRepository {
	return &requestRepo{db: db}
}

func (r *requestRepo) BeginTx(ctx context.Context) (*sqlx.Tx, error) {
	return r.db.BeginTxx(ctx, nil)
}

// Auction Requests
func (r *requestRepo) CreateAuctionRequest(ctx context.Context, req *models.AuctionRequest) error {
	query := `
		INSERT INTO auction_requests (
			user_id, category_id, location_id, title_ar, title_fr, title_en,
			description_ar, description_fr, description_en, start_price, min_increment,
			insurance_amount, reserve_price, buy_now_price, start_date, end_date,
			images, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
		RETURNING id, created_at
	`
	return r.db.QueryRowContext(ctx, query,
		req.UserID, req.CategoryID, req.LocationID, req.TitleAr, req.TitleFr, req.TitleEn,
		req.DescriptionAr, req.DescriptionFr, req.DescriptionEn, req.StartPrice, req.MinIncrement,
		req.InsuranceAmount, req.ReservePrice, req.BuyNowPrice, req.StartDate, req.EndDate,
		req.Images, req.Status,
	).Scan(&req.ID, &req.CreatedAt)
}

func (r *requestRepo) GetAuctionRequests(ctx context.Context, status string, userID *uuid.UUID, dateFrom, dateTo *time.Time, categoryID, locationID *int, minPrice, maxPrice *float64, sortBy, sortOrder string, page, perPage int) ([]models.AuctionRequest, int, error) {
	var requests []models.AuctionRequest
	offset := (page - 1) * perPage

	// Build WHERE clause dynamically
	whereConditions := []string{}
	args := []interface{}{}
	argIndex := 1

	if status != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("ar.status = $%d", argIndex))
		args = append(args, status)
		argIndex++
	}
	if userID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("ar.user_id = $%d", argIndex))
		args = append(args, *userID)
		argIndex++
	}
	if dateFrom != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("ar.created_at >= $%d", argIndex))
		args = append(args, *dateFrom)
		argIndex++
	}
	if dateTo != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("ar.created_at <= $%d", argIndex))
		args = append(args, *dateTo)
		argIndex++
	}
	if categoryID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("ar.category_id = $%d", argIndex))
		args = append(args, *categoryID)
		argIndex++
	}
	if locationID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("ar.location_id = $%d", argIndex))
		args = append(args, *locationID)
		argIndex++
	}
	if minPrice != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("ar.start_price >= $%d", argIndex))
		args = append(args, *minPrice)
		argIndex++
	}
	if maxPrice != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("ar.start_price <= $%d", argIndex))
		args = append(args, *maxPrice)
		argIndex++
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

 	validSortColumns := map[string]bool{
		"created_at": true,
		"start_date": true,
		"end_date": true,
		"start_price": true,
		"status": true,
	}
	validSortOrders := map[string]bool{
		"ASC": true,
		"DESC": true,
	}

	if sortBy == "" {
		sortBy = "created_at"
	}
	if sortOrder == "" {
		sortOrder = "DESC"
	}

	orderClause := "ORDER BY "
	if validSortColumns[sortBy] && validSortOrders[sortOrder] {
		orderClause += fmt.Sprintf("ar.%s %s", sortBy, sortOrder)
	} else {
		orderClause += "ar.created_at DESC"
	}

	// Get total count
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM auction_requests %s", whereClause)
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	// Get paginated results
	query := fmt.Sprintf(`
		SELECT
			ar.*,
			u.id as user_id, u.phone, u.full_name, u.role
		FROM auction_requests ar
		LEFT JOIN users u ON ar.user_id = u.id
		%s
		%s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderClause, argIndex, argIndex+1)
	args = append(args, perPage, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var req models.AuctionRequest
		var user models.User
		err := rows.Scan(
			&req.ID, &req.UserID, &req.CategoryID, &req.LocationID,
			&req.TitleAr, &req.TitleFr, &req.TitleEn,
			&req.DescriptionAr, &req.DescriptionFr, &req.DescriptionEn,
			&req.StartPrice, &req.MinIncrement, &req.InsuranceAmount,
			&req.ReservePrice, &req.BuyNowPrice, &req.StartDate, &req.EndDate,
			&req.Images, &req.Status, &req.AdminNotes, &req.ReviewedBy, &req.ReviewedAt,
			&req.CreatedAt, &req.UpdatedAt,
			&user.ID, &user.Phone, &user.FullName, &user.Role,
		)
		if err != nil {
			return nil, 0, err
		}
		req.User = &user
		requests = append(requests, req)
	}
	return requests, total, nil
}

func (r *requestRepo) GetAuctionRequestByID(ctx context.Context, id uuid.UUID) (*models.AuctionRequest, error) {
	var req models.AuctionRequest
	var user models.User
	query := `
		SELECT
			ar.*,
			u.id as user_id, u.phone, u.full_name, u.role
		FROM auction_requests ar
		LEFT JOIN users u ON ar.user_id = u.id
		WHERE ar.id = $1
	`
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&req.ID, &req.UserID, &req.CategoryID, &req.LocationID,
		&req.TitleAr, &req.TitleFr, &req.TitleEn,
		&req.DescriptionAr, &req.DescriptionFr, &req.DescriptionEn,
		&req.StartPrice, &req.MinIncrement, &req.InsuranceAmount,
		&req.ReservePrice, &req.BuyNowPrice, &req.StartDate, &req.EndDate,
		&req.Images, &req.Status, &req.AdminNotes, &req.ReviewedBy, &req.ReviewedAt,
		&req.CreatedAt, &req.UpdatedAt,
		&user.ID, &user.Phone, &user.FullName, &user.Role,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	req.User = &user
	return &req, nil
}

func (r *requestRepo) UpdateAuctionRequestStatus(ctx context.Context, id uuid.UUID, status, notes string, reviewedBy uuid.UUID) error {
	query := `
		UPDATE auction_requests
		SET status = $1, admin_notes = $2, reviewed_by = $3, reviewed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = $4
	`
	_, err := r.db.ExecContext(ctx, query, status, notes, reviewedBy, id)
	return err
}

func (r *requestRepo) UpdateAuctionRequestStatusTx(ctx context.Context, tx *sqlx.Tx, id uuid.UUID, status, notes string, reviewedBy uuid.UUID) error {
	query := `
		UPDATE auction_requests
		SET status = $1, admin_notes = $2, reviewed_by = $3, reviewed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = $4
	`
	_, err := tx.ExecContext(ctx, query, status, notes, reviewedBy, id)
	return err
}

func (r *requestRepo) DeleteAuctionRequest(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM auction_requests WHERE id = $1", id)
	return err
}

func (r *requestRepo) GetUserAuctionRequests(ctx context.Context, userID uuid.UUID, status string, page, perPage int) ([]models.AuctionRequest, int, error) {
	var requests []models.AuctionRequest
	offset := (page - 1) * perPage

	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM auction_requests WHERE user_id = $1 AND ($2 = '' OR status = $2)`
	if err := r.db.GetContext(ctx, &total, countQuery, userID, status); err != nil {
		return nil, 0, err
	}

	// Get paginated results
	query := `
		SELECT
			ar.*,
			u.id as user_id, u.phone, u.full_name, u.role
		FROM auction_requests ar
		LEFT JOIN users u ON ar.user_id = u.id
		WHERE ar.user_id = $1 AND ($2 = '' OR ar.status = $2)
		ORDER BY ar.created_at DESC
		LIMIT $3 OFFSET $4
	`
	rows, err := r.db.QueryContext(ctx, query, userID, status, perPage, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var req models.AuctionRequest
		var user models.User
		err := rows.Scan(
			&req.ID, &req.UserID, &req.CategoryID, &req.LocationID,
			&req.TitleAr, &req.TitleFr, &req.TitleEn,
			&req.DescriptionAr, &req.DescriptionFr, &req.DescriptionEn,
			&req.StartPrice, &req.MinIncrement, &req.InsuranceAmount,
			&req.ReservePrice, &req.BuyNowPrice, &req.StartDate, &req.EndDate,
			&req.Images, &req.Status, &req.AdminNotes, &req.ReviewedBy, &req.ReviewedAt,
			&req.CreatedAt, &req.UpdatedAt,
			&user.ID, &user.Phone, &user.FullName, &user.Role,
		)
		if err != nil {
			return nil, 0, err
		}
		req.User = &user
		requests = append(requests, req)
	}
	return requests, total, nil
}

// Banner Requests
func (r *requestRepo) CreateBannerRequest(ctx context.Context, req *models.BannerRequest) error {
	query := `
		INSERT INTO banner_requests (
			user_id, title_ar, title_fr, title_en, image_url, target_url,
			starts_at, ends_at, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at
	`
	return r.db.QueryRowContext(ctx, query,
		req.UserID, req.TitleAr, req.TitleFr, req.TitleEn, req.ImageURL, req.TargetURL,
		req.StartsAt, req.EndsAt, req.Status,
	).Scan(&req.ID, &req.CreatedAt)
}

func (r *requestRepo) GetBannerRequests(ctx context.Context, status string, userID *uuid.UUID, dateFrom, dateTo *time.Time, sortBy, sortOrder string, page, perPage int) ([]models.BannerRequest, int, error) {
	var requests []models.BannerRequest
	offset := (page - 1) * perPage

	// Build WHERE clause dynamically
	whereConditions := []string{}
	args := []interface{}{}
	argIndex := 1

	if status != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("br.status = $%d", argIndex))
		args = append(args, status)
		argIndex++
	}
	if userID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("br.user_id = $%d", argIndex))
		args = append(args, *userID)
		argIndex++
	}
	if dateFrom != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("br.created_at >= $%d", argIndex))
		args = append(args, *dateFrom)
		argIndex++
	}
	if dateTo != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("br.created_at <= $%d", argIndex))
		args = append(args, *dateTo)
		argIndex++
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Build ORDER BY clause dynamically
	validSortColumns := map[string]bool{
		"created_at": true,
		"starts_at": true,
		"ends_at": true,
		"status": true,
	}
	validSortOrders := map[string]bool{
		"ASC": true,
		"DESC": true,
	}

	if sortBy == "" {
		sortBy = "created_at"
	}
	if sortOrder == "" {
		sortOrder = "DESC"
	}

	orderClause := "ORDER BY "
	if validSortColumns[sortBy] && validSortOrders[sortOrder] {
		orderClause += fmt.Sprintf("br.%s %s", sortBy, sortOrder)
	} else {
		orderClause += "br.created_at DESC"
	}

	// Get total count
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM banner_requests %s", whereClause)
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	// Get paginated results
	query := fmt.Sprintf(`
		SELECT
			br.*,
			u.id as user_id, u.phone, u.full_name, u.role
		FROM banner_requests br
		LEFT JOIN users u ON br.user_id = u.id
		%s
		%s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderClause, argIndex, argIndex+1)
	args = append(args, perPage, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var req models.BannerRequest
		var user models.User
		err := rows.Scan(
			&req.ID, &req.UserID, &req.TitleAr, &req.TitleFr, &req.TitleEn,
			&req.ImageURL, &req.TargetURL, &req.StartsAt, &req.EndsAt,
			&req.Status, &req.AdminNotes, &req.ReviewedBy, &req.ReviewedAt,
			&req.CreatedAt, &req.UpdatedAt,
			&user.ID, &user.Phone, &user.FullName, &user.Role,
		)
		if err != nil {
			return nil, 0, err
		}
		req.User = &user
		requests = append(requests, req)
	}
	return requests, total, nil
}

func (r *requestRepo) GetBannerRequestByID(ctx context.Context, id uuid.UUID) (*models.BannerRequest, error) {
	var req models.BannerRequest
	var user models.User
	query := `
		SELECT
			br.*,
			u.id as user_id, u.phone, u.full_name, u.role
		FROM banner_requests br
		LEFT JOIN users u ON br.user_id = u.id
		WHERE br.id = $1
	`
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&req.ID, &req.UserID, &req.TitleAr, &req.TitleFr, &req.TitleEn,
		&req.ImageURL, &req.TargetURL, &req.StartsAt, &req.EndsAt,
		&req.Status, &req.AdminNotes, &req.ReviewedBy, &req.ReviewedAt,
		&req.CreatedAt, &req.UpdatedAt,
		&user.ID, &user.Phone, &user.FullName, &user.Role,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	req.User = &user
	return &req, nil
}

func (r *requestRepo) UpdateBannerRequestStatus(ctx context.Context, id uuid.UUID, status, notes string, reviewedBy uuid.UUID) error {
	query := `
		UPDATE banner_requests
		SET status = $1, admin_notes = $2, reviewed_by = $3, reviewed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = $4
	`
	_, err := r.db.ExecContext(ctx, query, status, notes, reviewedBy, id)
	return err
}

func (r *requestRepo) UpdateBannerRequestStatusTx(ctx context.Context, tx *sqlx.Tx, id uuid.UUID, status, notes string, reviewedBy uuid.UUID) error {
	query := `
		UPDATE banner_requests
		SET status = $1, admin_notes = $2, reviewed_by = $3, reviewed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = $4
	`
	_, err := tx.ExecContext(ctx, query, status, notes, reviewedBy, id)
	return err
}

func (r *requestRepo) DeleteBannerRequest(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM banner_requests WHERE id = $1", id)
	return err
}

func (r *requestRepo) GetUserBannerRequests(ctx context.Context, userID uuid.UUID, status string, page, perPage int) ([]models.BannerRequest, int, error) {
	var requests []models.BannerRequest
	offset := (page - 1) * perPage

	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM banner_requests WHERE user_id = $1 AND ($2 = '' OR status = $2)`
	if err := r.db.GetContext(ctx, &total, countQuery, userID, status); err != nil {
		return nil, 0, err
	}

	// Get paginated results
	query := `
		SELECT
			br.*,
			u.id as user_id, u.phone, u.full_name, u.role
		FROM banner_requests br
		LEFT JOIN users u ON br.user_id = u.id
		WHERE br.user_id = $1 AND ($2 = '' OR br.status = $2)
		ORDER BY br.created_at DESC
		LIMIT $3 OFFSET $4
	`
	rows, err := r.db.QueryContext(ctx, query, userID, status, perPage, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var req models.BannerRequest
		var user models.User
		err := rows.Scan(
			&req.ID, &req.UserID, &req.TitleAr, &req.TitleFr, &req.TitleEn,
			&req.ImageURL, &req.TargetURL, &req.StartsAt, &req.EndsAt,
			&req.Status, &req.AdminNotes, &req.ReviewedBy, &req.ReviewedAt,
			&req.CreatedAt, &req.UpdatedAt,
			&user.ID, &user.Phone, &user.FullName, &user.Role,
		)
		if err != nil {
			return nil, 0, err
		}
		req.User = &user
		requests = append(requests, req)
	}
	return requests, total, nil
}

func (r *requestRepo) CountPendingAuctionRequests(ctx context.Context) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM auction_requests WHERE status = 'pending'")
	return count, err
}

func (r *requestRepo) BulkUpdateAuctionRequestStatus(ctx context.Context, ids []uuid.UUID, status, notes string, reviewedBy uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}

	query := `
		UPDATE auction_requests
		SET status = $1, admin_notes = $2, reviewed_by = $3, reviewed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = ANY($4)
	`
	_, err := r.db.ExecContext(ctx, query, status, notes, reviewedBy, ids)
	return err
}

func (r *requestRepo) BulkDeleteAuctionRequests(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	query := `DELETE FROM auction_requests WHERE id = ANY($1)`
	_, err := r.db.ExecContext(ctx, query, ids)
	return err
}

func (r *requestRepo) CountPendingBannerRequests(ctx context.Context) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM banner_requests WHERE status = 'pending'")
	return count, err
}

func (r *requestRepo) BulkUpdateBannerRequestStatus(ctx context.Context, ids []uuid.UUID, status, notes string, reviewedBy uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}

	query := `
		UPDATE banner_requests
		SET status = $1, admin_notes = $2, reviewed_by = $3, reviewed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = ANY($4)
	`
	_, err := r.db.ExecContext(ctx, query, status, notes, reviewedBy, ids)
	return err
}

func (r *requestRepo) BulkDeleteBannerRequests(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	query := `DELETE FROM banner_requests WHERE id = ANY($1)`
	_, err := r.db.ExecContext(ctx, query, ids)
	return err
}
