package services

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/mazadpay/backend/internal/models"
)

type DeliveryDriverService interface {
	List(ctx context.Context) ([]models.DeliveryDriver, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.DeliveryDriver, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) (*models.DeliveryDriver, error)
	Create(ctx context.Context, driver *models.DeliveryDriver) error
	Update(ctx context.Context, id uuid.UUID, driver *models.DeliveryDriver) error
	Delete(ctx context.Context, id uuid.UUID) error
	UpdateLocation(ctx context.Context, id uuid.UUID, lat, lng float64) error
	UpdateAvailability(ctx context.Context, id uuid.UUID, isAvailable bool) error
}

type deliveryDriverService struct {
	db *sqlx.DB
}

func NewDeliveryDriverService(db *sqlx.DB) DeliveryDriverService {
	return &deliveryDriverService{db: db}
}

func (s *deliveryDriverService) List(ctx context.Context) ([]models.DeliveryDriver, error) {
	var drivers []models.DeliveryDriver
	err := s.db.SelectContext(ctx, &drivers, `
		SELECT id, user_id, vehicle_type, vehicle_plate, vehicle_color, 
		       license_number, rating, total_deliveries, is_available,
		       current_location_lat, current_location_lng, created_at
		FROM delivery_drivers
		ORDER BY created_at DESC
	`)
	return drivers, err
}

func (s *deliveryDriverService) GetByID(ctx context.Context, id uuid.UUID) (*models.DeliveryDriver, error) {
	var driver models.DeliveryDriver
	err := s.db.GetContext(ctx, &driver, `
		SELECT id, user_id, vehicle_type, vehicle_plate, vehicle_color,
		       license_number, rating, total_deliveries, is_available,
		       current_location_lat, current_location_lng, created_at
		FROM delivery_drivers WHERE id = $1
	`, id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("delivery driver not found")
	}
	return &driver, err
}

func (s *deliveryDriverService) GetByUserID(ctx context.Context, userID uuid.UUID) (*models.DeliveryDriver, error) {
	var driver models.DeliveryDriver
	err := s.db.GetContext(ctx, &driver, `
		SELECT id, user_id, vehicle_type, vehicle_plate, vehicle_color,
		       license_number, rating, total_deliveries, is_available,
		       current_location_lat, current_location_lng, created_at
		FROM delivery_drivers WHERE user_id = $1
	`, userID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("delivery driver not found for user")
	}
	return &driver, err
}

func (s *deliveryDriverService) Create(ctx context.Context, driver *models.DeliveryDriver) error {
	query := `
		INSERT INTO delivery_drivers 
		(user_id, vehicle_type, vehicle_plate, vehicle_color, license_number, is_available)
		VALUES (:user_id, :vehicle_type, :vehicle_plate, :vehicle_color, :license_number, :is_available)
		RETURNING id, created_at
	`
	rows, err := s.db.NamedQueryContext(ctx, query, driver)
	if err != nil {
		return err
	}
	defer rows.Close()
	
	if rows.Next() {
		return rows.Scan(&driver.ID, &driver.CreatedAt)
	}
	return nil
}

func (s *deliveryDriverService) Update(ctx context.Context, id uuid.UUID, driver *models.DeliveryDriver) error {
	query := `
		UPDATE delivery_drivers
		SET vehicle_type = :vehicle_type, vehicle_plate = :vehicle_plate,
		    vehicle_color = :vehicle_color, license_number = :license_number,
		    is_available = :is_available
		WHERE id = :id
	`
	driver.ID = id
	_, err := s.db.NamedExecContext(ctx, query, driver)
	return err
}

func (s *deliveryDriverService) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM delivery_drivers WHERE id = $1`, id)
	return err
}

func (s *deliveryDriverService) UpdateLocation(ctx context.Context, id uuid.UUID, lat, lng float64) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE delivery_drivers 
		SET current_location_lat = $1, current_location_lng = $2
		WHERE id = $3
	`, lat, lng, id)
	return err
}

func (s *deliveryDriverService) UpdateAvailability(ctx context.Context, id uuid.UUID, isAvailable bool) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE delivery_drivers 
		SET is_available = $1
		WHERE id = $2
	`, isAvailable, id)
	return err
}
