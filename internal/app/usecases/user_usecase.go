package usecases

import (
	"context"
	"fmt"

	"dental-scheduler-backend/internal/domain/entities"
	"dental-scheduler-backend/internal/domain/ports/repositories"
	"dental-scheduler-backend/internal/infra/logger"
)

// UserUseCase handles user-related business operations
type UserUseCase struct {
	userRepo repositories.UserRepository
	logger   *logger.Logger
}

// NewUserUseCase creates a new UserUseCase instance
func NewUserUseCase(userRepo repositories.UserRepository, logger *logger.Logger) *UserUseCase {
	return &UserUseCase{
		userRepo: userRepo,
		logger:   logger,
	}
}

// GetUserProfile retrieves a user's profile with organization information
func (u *UserUseCase) GetUserProfile(ctx context.Context, supabaseID string) (*entities.UserProfile, error) {
	u.logger.Logger.WithField("supabase_id", supabaseID).Info("Fetching user profile")

	profile, err := u.userRepo.GetProfileBySupabaseID(ctx, supabaseID)
	if err != nil {
		u.logger.Logger.WithError(err).Error("Failed to fetch user profile")
		return nil, fmt.Errorf("failed to fetch user profile: %w", err)
	}

	if profile == nil {
		u.logger.Logger.WithField("supabase_id", supabaseID).Warn("User profile not found")
		return nil, fmt.Errorf("user profile not found")
	}

	return profile, nil
}

// SyncUserFromSupabase creates or updates a user based on Supabase authentication data
func (u *UserUseCase) SyncUserFromSupabase(ctx context.Context, supabaseID, email string) (*entities.Profile, error) {
	u.logger.Logger.WithFields(map[string]interface{}{
		"supabase_id": supabaseID,
		"email":       email,
	}).Info("Syncing user from Supabase")

	// Try to find existing profile
	existingProfile, err := u.userRepo.GetBySupabaseID(ctx, supabaseID)
	if err != nil {
		// If profile doesn't exist, we might want to create them
		// This depends on your business logic
		u.logger.Logger.WithError(err).Debug("Profile not found in local database")
	}

	if existingProfile != nil {
		return existingProfile, nil
	}

	// If profile doesn't exist, you might want to handle this case
	// based on your business requirements
	return nil, fmt.Errorf("profile not found in local database: %s", supabaseID)
}
