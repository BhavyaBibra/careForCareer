package readiness

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, ra *ReadinessAssessment) error
	GetByID(ctx context.Context, id uuid.UUID) (*ReadinessAssessment, error)
	GetLatestByCandidateAndJD(ctx context.Context, candidateID, jdID uuid.UUID) (*ReadinessAssessment, error)
	ListByCandidateID(ctx context.Context, candidateID uuid.UUID, limit int, cursor string) ([]*ReadinessAssessment, string, error)
}
