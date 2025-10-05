package entities

import (
	"time"

	"github.com/google/uuid"
)

// Service represents a dental service that can be offered by the organization
type Service struct {
	ID             string
	Name           string
	BasePrice      *float64
	OrganizationID uuid.UUID
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
