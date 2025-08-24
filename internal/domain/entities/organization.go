package entities

import (
	"time"

	"github.com/google/uuid"
)

// Organization represents a dental organization that can have multiple clinics
type Organization struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description *string   `json:"description,omitempty" db:"description"`
	Address     *string   `json:"address,omitempty" db:"address"`
	Phone       *string   `json:"phone,omitempty" db:"phone"`
	Email       *string   `json:"email,omitempty" db:"email"`
	Website     *string   `json:"website,omitempty" db:"website"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// NewOrganization creates a new organization with the given name
func NewOrganization(name string) *Organization {
	now := time.Now()
	return &Organization{
		ID:        uuid.New(),
		Name:      name,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Validate validates the organization fields
func (o *Organization) Validate() error {
	if o.Name == "" {
		return ErrInvalidOrganizationName
	}
	return nil
}

// SetDescription sets the organization description
func (o *Organization) SetDescription(description string) {
	o.Description = &description
	o.UpdatedAt = time.Now()
}

// SetAddress sets the organization address
func (o *Organization) SetAddress(address string) {
	o.Address = &address
	o.UpdatedAt = time.Now()
}

// SetContact sets the organization contact information
func (o *Organization) SetContact(phone, email, website string) {
	if phone != "" {
		o.Phone = &phone
	}
	if email != "" {
		o.Email = &email
	}
	if website != "" {
		o.Website = &website
	}
	o.UpdatedAt = time.Now()
}

// Deactivate deactivates the organization
func (o *Organization) Deactivate() {
	o.IsActive = false
	o.UpdatedAt = time.Now()
}

// Activate activates the organization
func (o *Organization) Activate() {
	o.IsActive = true
	o.UpdatedAt = time.Now()
}
