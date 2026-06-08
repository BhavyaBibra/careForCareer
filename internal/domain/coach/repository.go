package coach

import (
	"context"

	"github.com/google/uuid"
)

type SessionRepository interface {
	Create(ctx context.Context, s *CoachSession) error
	GetByID(ctx context.Context, id uuid.UUID) (*CoachSession, error)
	ListByCandidateID(ctx context.Context, candidateID uuid.UUID, limit int, cursor string) ([]*CoachSession, string, error)
}

type MessageRepository interface {
	Create(ctx context.Context, m *CoachMessage) error
	ListBySessionID(ctx context.Context, sessionID uuid.UUID) ([]*CoachMessage, error)
	CountTodayByUserID(ctx context.Context, userID uuid.UUID) (int, error)
}
