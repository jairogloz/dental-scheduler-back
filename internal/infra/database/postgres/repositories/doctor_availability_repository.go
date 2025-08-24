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

// DoctorAvailabilityPostgresRepository implements the DoctorAvailabilityRepository interface
type DoctorAvailabilityPostgresRepository struct {
	db *sql.DB
}

// NewDoctorAvailabilityPostgresRepository creates a new instance of DoctorAvailabilityPostgresRepository
func NewDoctorAvailabilityPostgresRepository(db *sql.DB) repositories.DoctorAvailabilityRepository {
	return &DoctorAvailabilityPostgresRepository{db: db}
}

// Create creates a new doctor availability entry
func (r *DoctorAvailabilityPostgresRepository) Create(ctx context.Context, availability *entities.DoctorAvailability) error {
	query := `
		INSERT INTO doctor_availability (id, doctor_id, start_time, end_time, recurrence_rule, is_available, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.db.ExecContext(ctx, query,
		availability.ID,
		availability.DoctorID,
		availability.StartTime,
		availability.EndTime,
		availability.RecurrenceRule,
		availability.IsAvailable,
		availability.CreatedAt,
		availability.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create doctor availability: %w", err)
	}

	return nil
}

// GetByID retrieves a doctor availability by its ID
func (r *DoctorAvailabilityPostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.DoctorAvailability, error) {
	query := `
		SELECT id, doctor_id, start_time, end_time, recurrence_rule, is_available, created_at, updated_at
		FROM doctor_availability
		WHERE id = $1`

	var availability entities.DoctorAvailability
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&availability.ID,
		&availability.DoctorID,
		&availability.StartTime,
		&availability.EndTime,
		&availability.RecurrenceRule,
		&availability.IsAvailable,
		&availability.CreatedAt,
		&availability.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get doctor availability: %w", err)
	}

	return &availability, nil
}

// GetByDoctorID retrieves all availability entries for a doctor
func (r *DoctorAvailabilityPostgresRepository) GetByDoctorID(ctx context.Context, doctorID uuid.UUID) ([]*entities.DoctorAvailability, error) {
	query := `
		SELECT id, doctor_id, start_time, end_time, recurrence_rule, is_available, created_at, updated_at
		FROM doctor_availability
		WHERE doctor_id = $1
		ORDER BY start_time`

	rows, err := r.db.QueryContext(ctx, query, doctorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get doctor availability by doctor ID: %w", err)
	}
	defer rows.Close()

	return r.scanAvailabilities(rows)
}

// GetByDoctorIDAndDate retrieves availability for a doctor on a specific date
func (r *DoctorAvailabilityPostgresRepository) GetByDoctorIDAndDate(ctx context.Context, doctorID uuid.UUID, date time.Time) ([]*entities.DoctorAvailability, error) {
	// Get the start and end of the day
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	query := `
		SELECT id, doctor_id, start_time, end_time, recurrence_rule, is_available, created_at, updated_at
		FROM doctor_availability
		WHERE doctor_id = $1 
		  AND start_time < $3 
		  AND end_time > $2
		ORDER BY start_time`

	rows, err := r.db.QueryContext(ctx, query, doctorID, startOfDay, endOfDay)
	if err != nil {
		return nil, fmt.Errorf("failed to get doctor availability by doctor ID and date: %w", err)
	}
	defer rows.Close()

	return r.scanAvailabilities(rows)
}

// Update updates an existing doctor availability
func (r *DoctorAvailabilityPostgresRepository) Update(ctx context.Context, availability *entities.DoctorAvailability) error {
	query := `
		UPDATE doctor_availability
		SET start_time = $2, end_time = $3, recurrence_rule = $4, is_available = $5, updated_at = $6
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query,
		availability.ID,
		availability.StartTime,
		availability.EndTime,
		availability.RecurrenceRule,
		availability.IsAvailable,
		availability.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update doctor availability: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return entities.ErrAvailabilityNotFound
	}

	return nil
}

// Delete deletes a doctor availability by its ID
func (r *DoctorAvailabilityPostgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM doctor_availability WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete doctor availability: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return entities.ErrAvailabilityNotFound
	}

	return nil
}

// IsAvailable checks if a doctor is available during a specific time range
func (r *DoctorAvailabilityPostgresRepository) IsAvailable(ctx context.Context, doctorID uuid.UUID, startTime, endTime time.Time) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM doctor_availability
		WHERE doctor_id = $1
		  AND is_available = true
		  AND start_time <= $2
		  AND end_time >= $3`

	var count int
	err := r.db.QueryRowContext(ctx, query, doctorID, startTime, endTime).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check doctor availability: %w", err)
	}

	return count > 0, nil
}

// scanAvailabilities is a helper method to scan multiple availability rows
func (r *DoctorAvailabilityPostgresRepository) scanAvailabilities(rows *sql.Rows) ([]*entities.DoctorAvailability, error) {
	var availabilities []*entities.DoctorAvailability
	for rows.Next() {
		var availability entities.DoctorAvailability
		err := rows.Scan(
			&availability.ID,
			&availability.DoctorID,
			&availability.StartTime,
			&availability.EndTime,
			&availability.RecurrenceRule,
			&availability.IsAvailable,
			&availability.CreatedAt,
			&availability.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan doctor availability: %w", err)
		}
		availabilities = append(availabilities, &availability)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over availability rows: %w", err)
	}

	return availabilities, nil
}
