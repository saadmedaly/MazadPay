package services

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/mazadpay/backend/internal/models"
)

type PaymentMethodService interface {
	List(ctx context.Context) ([]models.PaymentMethod, error)
	GetByID(ctx context.Context, id int) (*models.PaymentMethod, error)
	Create(ctx context.Context, pm *models.PaymentMethod) error
	Update(ctx context.Context, id int, pm *models.PaymentMethod) error
	Delete(ctx context.Context, id int) error
	ToggleStatus(ctx context.Context, id int) error
}

type paymentMethodService struct {
	db *sqlx.DB
}

func NewPaymentMethodService(db *sqlx.DB) PaymentMethodService {
	return &paymentMethodService{db: db}
}

func (s *paymentMethodService) List(ctx context.Context) ([]models.PaymentMethod, error) {
	var methods []models.PaymentMethod
	err := s.db.SelectContext(ctx, &methods, `
		SELECT id, code, name_ar, name_fr, name_en, logo_url, is_active, country_id, created_at
		FROM payment_methods
		ORDER BY id
	`)
	return methods, err
}

func (s *paymentMethodService) GetByID(ctx context.Context, id int) (*models.PaymentMethod, error) {
	var pm models.PaymentMethod
	err := s.db.GetContext(ctx, &pm, `
		SELECT id, code, name_ar, name_fr, name_en, logo_url, is_active, country_id, created_at
		FROM payment_methods WHERE id = $1
	`, id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("payment method not found")
	}
	return &pm, err
}

func (s *paymentMethodService) Create(ctx context.Context, pm *models.PaymentMethod) error {
	query := `
		INSERT INTO payment_methods (code, name_ar, name_fr, name_en, logo_url, is_active, country_id)
		VALUES (:code, :name_ar, :name_fr, :name_en, :logo_url, :is_active, :country_id)
		RETURNING id, created_at
	`
	rows, err := s.db.NamedQueryContext(ctx, query, pm)
	if err != nil {
		return err
	}
	defer rows.Close()
	
	if rows.Next() {
		return rows.Scan(&pm.ID, &pm.CreatedAt)
	}
	return nil
}

func (s *paymentMethodService) Update(ctx context.Context, id int, pm *models.PaymentMethod) error {
	query := `
		UPDATE payment_methods
		SET code = :code, name_ar = :name_ar, name_fr = :name_fr, 
		    name_en = :name_en, logo_url = :logo_url, 
		    is_active = :is_active, country_id = :country_id
		WHERE id = :id
	`
	pm.ID = id
	_, err := s.db.NamedExecContext(ctx, query, pm)
	return err
}

func (s *paymentMethodService) Delete(ctx context.Context, id int) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM payment_methods WHERE id = $1`, id)
	return err
}

func (s *paymentMethodService) ToggleStatus(ctx context.Context, id int) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE payment_methods 
		SET is_active = NOT is_active 
		WHERE id = $1
	`, id)
	return err
}
