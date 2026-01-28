package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"dental-scheduler-backend/internal/domain/entities"
	"dental-scheduler-backend/internal/domain/ports/repositories"

	"github.com/google/uuid"
)

// AppointmentAccountPostgresRepository implements the AppointmentAccountRepository interface
type AppointmentAccountPostgresRepository struct {
	db *sql.DB
}

// NewAppointmentAccountPostgresRepository creates a new instance of AppointmentAccountPostgresRepository
func NewAppointmentAccountPostgresRepository(db *sql.DB) repositories.AppointmentAccountRepository {
	return &AppointmentAccountPostgresRepository{db: db}
}

// Create creates a new appointment account
func (r *AppointmentAccountPostgresRepository) Create(ctx context.Context, account *entities.AppointmentAccount) error {
	query := `
		INSERT INTO appointment_accounts (id, organization_id, appointment_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)`

	_, err := r.db.ExecContext(ctx, query,
		account.ID,
		account.OrganizationID,
		account.AppointmentID,
		account.CreatedAt,
		account.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create appointment account: %w", err)
	}

	return nil
}

// GetByID retrieves an appointment account by its ID
func (r *AppointmentAccountPostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.AppointmentAccount, error) {
	query := `
		SELECT id, organization_id, appointment_id, created_at, updated_at
		FROM appointment_accounts
		WHERE id = $1`

	var account entities.AppointmentAccount
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&account.ID,
		&account.OrganizationID,
		&account.AppointmentID,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get appointment account: %w", err)
	}

	return &account, nil
}

// GetByAppointmentID retrieves an appointment account by appointment ID
func (r *AppointmentAccountPostgresRepository) GetByAppointmentID(ctx context.Context, appointmentID uuid.UUID) (*entities.AppointmentAccount, error) {
	query := `
		SELECT id, organization_id, appointment_id, created_at, updated_at
		FROM appointment_accounts
		WHERE appointment_id = $1`

	var account entities.AppointmentAccount
	err := r.db.QueryRowContext(ctx, query, appointmentID).Scan(
		&account.ID,
		&account.OrganizationID,
		&account.AppointmentID,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get appointment account by appointment ID: %w", err)
	}

	return &account, nil
}

// Exists checks if an appointment account exists for an appointment
func (r *AppointmentAccountPostgresRepository) Exists(ctx context.Context, appointmentID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM appointment_accounts WHERE appointment_id = $1)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, appointmentID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check appointment account existence: %w", err)
	}

	return exists, nil
}

// GetOrCreate gets an existing account or creates a new one if it doesn't exist
func (r *AppointmentAccountPostgresRepository) GetOrCreate(ctx context.Context, organizationID, appointmentID uuid.UUID) (*entities.AppointmentAccount, error) {
	// First try to get existing account
	account, err := r.GetByAppointmentID(ctx, appointmentID)
	if err != nil {
		return nil, err
	}

	if account != nil {
		return account, nil
	}

	// Create new account if it doesn't exist
	newAccount := &entities.AppointmentAccount{
		ID:             uuid.New(),
		OrganizationID: organizationID,
		AppointmentID:  appointmentID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	err = r.Create(ctx, newAccount)
	if err != nil {
		return nil, err
	}

	return newAccount, nil
}
