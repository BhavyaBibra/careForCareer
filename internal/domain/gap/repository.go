package gap

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, ga *GapAnalysis) error
	GetByID(ctx context.Context, id uuid.UUID) (*GapAnalysis, error)
	GetLatestByCandidateAndJD(ctx context.Context, candidateID, jdID uuid.UUID) (*GapAnalysis, error)
}
