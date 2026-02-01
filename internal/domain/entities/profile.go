package entities

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// Role represents user roles in the system
type Role string

const (
	RoleAdmin        Role = "admin"
	RoleDoctor       Role = "doctor"
	RoleReceptionist Role = "receptionist"
	RolePatient      Role = "patient"
	RoleDev          Role = "dev"
)

// Profile represents a user profile linked to Supabase auth
type Profile struct {
	ID              uuid.UUID      `json:"id" db:"id"`                               // Links to auth.users(id)
	OrganizationID  *uuid.UUID     `json:"organization_id" db:"organization_id"`     // Optional organization link
	DefaultClinicID *uuid.UUID     `json:"default_clinic_id" db:"default_clinic_id"` // Optional default clinic
	Email           string         `json:"email" db:"email"`
	FullName        *string        `json:"full_name,omitempty" db:"full_name"`
	Roles           pq.StringArray `json:"roles" db:"roles"`
	AvatarURL       *string        `json:"avatar_url,omitempty" db:"avatar_url"`
	CreatedAt       time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at" db:"updated_at"`
}

// NewProfile creates a new profile with default receptionist role
func NewProfile(id uuid.UUID, email string) *Profile {
	now := time.Now()
	return &Profile{
		ID:        id,
		Email:     email,
		Roles:     pq.StringArray{string(RoleReceptionist)},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Validate validates the profile fields
func (p *Profile) Validate() error {
	if p.ID == uuid.Nil {
		return ErrInvalidProfileID
	}
	if p.Email == "" {
		return ErrInvalidEmail
	}
	if len(p.Roles) == 0 {
		return ErrInvalidRoles
	}
	return nil
}

// HasRole checks if the profile has a specific role
func (p *Profile) HasRole(role Role) bool {
	for _, r := range p.Roles {
		if Role(r) == role {
			return true
		}
	}
	return false
}

// AddRole adds a role to the profile if not already present
func (p *Profile) AddRole(role Role) {
	if !p.HasRole(role) {
		p.Roles = append(p.Roles, string(role))
		p.UpdatedAt = time.Now()
	}
}

// RemoveRole removes a role from the profile
func (p *Profile) RemoveRole(role Role) {
	for i, r := range p.Roles {
		if Role(r) == role {
			p.Roles = append(p.Roles[:i], p.Roles[i+1:]...)
			p.UpdatedAt = time.Now()
			break
		}
	}
}

// IsAdmin checks if the profile has admin role
func (p *Profile) IsAdmin() bool {
	return p.HasRole(RoleAdmin)
}

// IsDoctor checks if the profile has doctor role
func (p *Profile) IsDoctor() bool {
	return p.HasRole(RoleDoctor)
}

// SetOrganization sets the organization for the profile
func (p *Profile) SetOrganization(organizationID uuid.UUID) {
	p.OrganizationID = &organizationID
	p.UpdatedAt = time.Now()
}

// SetDefaultClinic sets the default clinic for the profile
func (p *Profile) SetDefaultClinic(clinicID uuid.UUID) {
	p.DefaultClinicID = &clinicID
	p.UpdatedAt = time.Now()
}

// ClearDefaultClinic removes the default clinic from the profile
func (p *Profile) ClearDefaultClinic() {
	p.DefaultClinicID = nil
	p.UpdatedAt = time.Now()
}

// UpdateProfile updates profile information
func (p *Profile) UpdateProfile(fullName, avatarURL *string) {
	if fullName != nil {
		p.FullName = fullName
	}
	if avatarURL != nil {
		p.AvatarURL = avatarURL
	}
	p.UpdatedAt = time.Now()
}

// UserProfile represents user profile information with organization details
type UserProfile struct {
	Profile      *Profile      `json:"profile"`
	Organization *Organization `json:"organization"`
}
