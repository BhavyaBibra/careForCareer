package roadmap

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, r *Roadmap) error
	GetByAssessmentID(ctx context.Context, assessmentID uuid.UUID) (*Roadmap, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Roadmap, error)
	Update(ctx context.Context, r *Roadmap) error
}
