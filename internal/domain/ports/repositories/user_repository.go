package repositories

import (
	"context"
	"dental-scheduler-backend/internal/domain/entities"

	"github.com/google/uuid"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	// GetByID retrieves a profile by their ID
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Profile, error)

	// GetByEmail retrieves a profile by their email
	GetByEmail(ctx context.Context, email string) (*entities.Profile, error)

	// GetBySupabaseID retrieves a profile by their Supabase user ID
	GetBySupabaseID(ctx context.Context, supabaseID string) (*entities.Profile, error)

	// GetProfileBySupabaseID retrieves a user profile with organization info by Supabase ID
	GetProfileBySupabaseID(ctx context.Context, supabaseID string) (*entities.UserProfile, error)

	// Create creates a new profile
	Create(ctx context.Context, profile *entities.Profile) error

	// Update updates an existing profile
	Update(ctx context.Context, profile *entities.Profile) error
}
