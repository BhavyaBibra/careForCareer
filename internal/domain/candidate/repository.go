package candidate

import (
	"context"

	"github.com/google/uuid"
)

// Repository is the port (interface) for candidate persistence.
// Implemented in infrastructure/postgres.
type Repository interface {
	Create(ctx context.Context, c *Candidate) error
	GetByID(ctx context.Context, id uuid.UUID) (*Candidate, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) (*Candidate, error)
	Update(ctx context.Context, c *Candidate) error
	Delete(ctx context.Context, id uuid.UUID) error
}
