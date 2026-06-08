package resume

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, r *Resume) error
	GetByID(ctx context.Context, id uuid.UUID) (*Resume, error)
	GetLatestByCandidateID(ctx context.Context, candidateID uuid.UUID) (*Resume, error)
	Update(ctx context.Context, r *Resume) error
	NextVersion(ctx context.Context, candidateID uuid.UUID) (int, error)
}
