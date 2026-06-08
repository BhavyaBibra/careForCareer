package jd

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, j *JobDescription) error
	GetByID(ctx context.Context, id uuid.UUID) (*JobDescription, error)
	Update(ctx context.Context, j *JobDescription) error
	ListByCandidateID(ctx context.Context, candidateID uuid.UUID, limit int, cursor string) ([]*JobDescription, string, error)
}

type JDSkillRepository interface {
	BulkCreate(ctx context.Context, skills []*JDSkill) error
	ListByJDID(ctx context.Context, jdID uuid.UUID) ([]*JDSkill, error)
	DeleteByJDID(ctx context.Context, jdID uuid.UUID) error
}
