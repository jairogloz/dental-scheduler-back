package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"dental-scheduler-backend/internal/domain/entities"
	"dental-scheduler-backend/internal/domain/ports/repositories"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// UserPostgresRepository implements the UserRepository interface
type UserPostgresRepository struct {
	db *sql.DB
}

// NewUserPostgresRepository creates a new instance of UserPostgresRepository
func NewUserPostgresRepository(db *sql.DB) repositories.UserRepository {
	return &UserPostgresRepository{db: db}
}

// GetByID retrieves a profile by ID
func (r *UserPostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Profile, error) {
	query := `
		SELECT id, email, full_name, roles, organization_id, avatar_url, created_at, updated_at
		FROM profiles
		WHERE id = $1`

	profile := &entities.Profile{}
	var fullName, avatarURL sql.NullString
	var organizationID sql.NullString
	var roles pq.StringArray

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&profile.ID,
		&profile.Email,
		&fullName,
		&roles,
		&organizationID,
		&avatarURL,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("profile not found")
		}
		return nil, fmt.Errorf("failed to get profile by ID: %w", err)
	}

	// Assign scanned roles to profile
	profile.Roles = roles

	// Handle nullable fields
	if fullName.Valid {
		profile.FullName = &fullName.String
	}
	if avatarURL.Valid {
		profile.AvatarURL = &avatarURL.String
	}
	if organizationID.Valid {
		orgUUID, err := uuid.Parse(organizationID.String)
		if err != nil {
			return nil, fmt.Errorf("invalid organization ID format: %w", err)
		}
		profile.OrganizationID = &orgUUID
	}

	return profile, nil
}

// GetByEmail retrieves a profile by email
func (r *UserPostgresRepository) GetByEmail(ctx context.Context, email string) (*entities.Profile, error) {
	query := `
		SELECT id, email, full_name, roles, organization_id, avatar_url, created_at, updated_at
		FROM profiles
		WHERE email = $1`

	profile := &entities.Profile{}
	var fullName, avatarURL sql.NullString
	var organizationID sql.NullString

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&profile.ID,
		&profile.Email,
		&fullName,
		pq.Array(&profile.Roles),
		&organizationID,
		&avatarURL,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("profile not found")
		}
		return nil, fmt.Errorf("failed to get profile by email: %w", err)
	}

	// Handle nullable fields
	if fullName.Valid {
		profile.FullName = &fullName.String
	}
	if avatarURL.Valid {
		profile.AvatarURL = &avatarURL.String
	}
	if organizationID.Valid {
		orgUUID, err := uuid.Parse(organizationID.String)
		if err != nil {
			return nil, fmt.Errorf("invalid organization ID format: %w", err)
		}
		profile.OrganizationID = &orgUUID
	}

	return profile, nil
}

// GetBySupabaseID retrieves a profile by their Supabase UUID (stored as string)
func (r *UserPostgresRepository) GetBySupabaseID(ctx context.Context, supabaseID string) (*entities.Profile, error) {
	// Supabase ID is the same as the profile ID in the profiles table
	profileUUID, err := uuid.Parse(supabaseID)
	if err != nil {
		return nil, fmt.Errorf("invalid supabase ID format: %w", err)
	}

	return r.GetByID(ctx, profileUUID)
}

// GetProfileBySupabaseID retrieves user profile with organization details
func (r *UserPostgresRepository) GetProfileBySupabaseID(ctx context.Context, supabaseID string) (*entities.UserProfile, error) {
	query := `
		SELECT 
			p.id, p.email, p.full_name, p.roles, p.organization_id, p.avatar_url, p.created_at, p.updated_at,
			o.id, o.name, o.description, o.address, o.phone, o.email, o.is_active, o.created_at, o.updated_at
		FROM profiles p
		LEFT JOIN organizations o ON p.organization_id = o.id
		WHERE p.id = $1`

	profileUUID, err := uuid.Parse(supabaseID)
	if err != nil {
		return nil, fmt.Errorf("invalid supabase ID format: %w", err)
	}

	profile := &entities.Profile{}
	var fullName, avatarURL sql.NullString
	var profileOrgID sql.NullString
	var orgID, orgName, orgDescription, orgAddress, orgPhone, orgEmail sql.NullString
	var orgIsActive sql.NullBool
	var orgCreatedAt, orgUpdatedAt sql.NullTime
	var roles pq.StringArray

	err = r.db.QueryRowContext(ctx, query, profileUUID).Scan(
		&profile.ID,
		&profile.Email,
		&fullName,
		&roles,
		&profileOrgID,
		&avatarURL,
		&profile.CreatedAt,
		&profile.UpdatedAt,
		&orgID,
		&orgName,
		&orgDescription,
		&orgAddress,
		&orgPhone,
		&orgEmail,
		&orgIsActive,
		&orgCreatedAt,
		&orgUpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user profile not found")
		}
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	// Assign scanned roles to profile
	profile.Roles = roles

	// Handle profile nullable fields
	if fullName.Valid {
		profile.FullName = &fullName.String
	}
	if avatarURL.Valid {
		profile.AvatarURL = &avatarURL.String
	}
	if profileOrgID.Valid {
		orgUUID, err := uuid.Parse(profileOrgID.String)
		if err != nil {
			return nil, fmt.Errorf("invalid profile organization ID format: %w", err)
		}
		profile.OrganizationID = &orgUUID
	}

	// Build organization if data exists
	var org *entities.Organization
	if orgID.Valid {
		orgUUID, err := uuid.Parse(orgID.String)
		if err != nil {
			return nil, fmt.Errorf("invalid organization ID format: %w", err)
		}

		org = &entities.Organization{
			ID:       orgUUID,
			Name:     orgName.String,
			IsActive: orgIsActive.Bool,
		}

		// Set optional fields if they have values
		if orgDescription.Valid {
			org.Description = &orgDescription.String
		}
		if orgAddress.Valid {
			org.Address = &orgAddress.String
		}
		if orgPhone.Valid {
			org.Phone = &orgPhone.String
		}
		if orgEmail.Valid {
			org.Email = &orgEmail.String
		}
		if orgCreatedAt.Valid {
			org.CreatedAt = orgCreatedAt.Time
		}
		if orgUpdatedAt.Valid {
			org.UpdatedAt = orgUpdatedAt.Time
		}
	}

	userProfile := &entities.UserProfile{
		Profile:      profile,
		Organization: org,
	}

	return userProfile, nil
}

// Create creates a new profile
func (r *UserPostgresRepository) Create(ctx context.Context, profile *entities.Profile) error {
	query := `
		INSERT INTO profiles (id, email, full_name, roles, organization_id, avatar_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.db.ExecContext(ctx, query,
		profile.ID,
		profile.Email,
		profile.FullName,
		pq.Array(profile.Roles),
		profile.OrganizationID,
		profile.AvatarURL,
		profile.CreatedAt,
		profile.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create profile: %w", err)
	}

	return nil
}

// Update updates an existing profile
func (r *UserPostgresRepository) Update(ctx context.Context, profile *entities.Profile) error {
	query := `
		UPDATE profiles 
		SET email = $2, full_name = $3, roles = $4, organization_id = $5, avatar_url = $6, updated_at = $7
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query,
		profile.ID,
		profile.Email,
		profile.FullName,
		pq.Array(profile.Roles),
		profile.OrganizationID,
		profile.AvatarURL,
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to update profile: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("profile not found")
	}

	return nil
}
