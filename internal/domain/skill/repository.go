package skill

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*Skill, error)
	GetByName(ctx context.Context, name string) (*Skill, error)
	ResolveAlias(ctx context.Context, alias string) (*Skill, error) // resolves alias OR name to canonical skill
	ListAll(ctx context.Context) ([]*Skill, error)
	Create(ctx context.Context, s *Skill) error
}

type CandidateSkillRepository interface {
	Upsert(ctx context.Context, cs *CandidateSkill) error
	ListByCandidateID(ctx context.Context, candidateID uuid.UUID) ([]*CandidateSkill, error)
	ListByCandidateAndResume(ctx context.Context, candidateID, resumeID uuid.UUID) ([]*CandidateSkill, error)
	DeleteByResumeID(ctx context.Context, resumeID uuid.UUID) error
}
