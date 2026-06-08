package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"careergps/internal/domain/skill"
)

type SkillRepo struct {
	pool *pgxpool.Pool
}

func NewSkillRepo(pool *pgxpool.Pool) *SkillRepo {
	return &SkillRepo{pool: pool}
}

func (r *SkillRepo) GetByID(ctx context.Context, id uuid.UUID) (*skill.Skill, error) {
	const q = `SELECT id, name, category, aliases, created_at, updated_at
		FROM skills WHERE id = $1`
	row := r.pool.QueryRow(ctx, q, id)
	return scanSkill(row)
}

func (r *SkillRepo) GetByName(ctx context.Context, name string) (*skill.Skill, error) {
	const q = `SELECT id, name, category, aliases, created_at, updated_at
		FROM skills WHERE name = $1`
	row := r.pool.QueryRow(ctx, q, name)
	return scanSkill(row)
}

func (r *SkillRepo) ResolveAlias(ctx context.Context, alias string) (*skill.Skill, error) {
	const q = `SELECT id, name, category, aliases, created_at, updated_at
		FROM skills WHERE name = $1 OR $1 = ANY(aliases)`
	row := r.pool.QueryRow(ctx, q, alias)
	return scanSkill(row)
}

func (r *SkillRepo) ListAll(ctx context.Context) ([]*skill.Skill, error) {
	const q = `SELECT id, name, category, aliases, created_at, updated_at
		FROM skills ORDER BY name`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("skill_repo: list_all: %w", err)
	}
	defer rows.Close()

	var out []*skill.Skill
	for rows.Next() {
		s, err := scanSkill(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *SkillRepo) Create(ctx context.Context, s *skill.Skill) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	now := time.Now().UTC()
	s.CreatedAt = now
	s.UpdatedAt = now
	const q = `INSERT INTO skills (id, name, category, aliases, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.pool.Exec(ctx, q, s.ID, s.Name, string(s.Category), s.Aliases, s.CreatedAt, s.UpdatedAt)
	if err != nil {
		return fmt.Errorf("skill_repo: create: %w", err)
	}
	return nil
}

// scanSkill works on both pgx.Row and pgx.Rows via the RowScanner interface.
func scanSkill(row interface {
	Scan(dest ...any) error
}) (*skill.Skill, error) {
	var s skill.Skill
	var cat string
	var aliases []string
	err := row.Scan(&s.ID, &s.Name, &cat, &aliases, &s.CreatedAt, &s.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("skill_repo: not found")
	}
	if err != nil {
		return nil, fmt.Errorf("skill_repo: scan: %w", err)
	}
	s.Category = skill.Category(cat)
	s.Aliases = aliases
	return &s, nil
}

// CandidateSkillRepo implements skill.CandidateSkillRepository.
type CandidateSkillRepo struct {
	pool *pgxpool.Pool
}

func NewCandidateSkillRepo(pool *pgxpool.Pool) *CandidateSkillRepo {
	return &CandidateSkillRepo{pool: pool}
}

func (r *CandidateSkillRepo) Upsert(ctx context.Context, cs *skill.CandidateSkill) error {
	if cs.ID == uuid.Nil {
		cs.ID = uuid.New()
	}
	cs.CreatedAt = time.Now().UTC()
	const q = `INSERT INTO candidate_skills
		(id, candidate_id, skill_id, resume_id, score, confidence, evidence_source, raw_evidence, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (candidate_id, skill_id, resume_id)
		DO UPDATE SET score=$5, confidence=$6, evidence_source=$7, raw_evidence=$8`
	_, err := r.pool.Exec(ctx, q,
		cs.ID, cs.CandidateID, cs.SkillID, cs.ResumeID,
		cs.Score, cs.Confidence, cs.EvidenceSource, cs.RawEvidence, cs.CreatedAt)
	if err != nil {
		return fmt.Errorf("candidate_skill_repo: upsert: %w", err)
	}
	return nil
}

func (r *CandidateSkillRepo) ListByCandidateID(ctx context.Context, candidateID uuid.UUID) ([]*skill.CandidateSkill, error) {
	const q = `SELECT id, candidate_id, skill_id, resume_id, score, confidence, evidence_source, raw_evidence, created_at
		FROM candidate_skills WHERE candidate_id=$1 ORDER BY score DESC`
	rows, err := r.pool.Query(ctx, q, candidateID)
	if err != nil {
		return nil, fmt.Errorf("candidate_skill_repo: list_by_candidate: %w", err)
	}
	defer rows.Close()
	return scanCandidateSkills(rows)
}

func (r *CandidateSkillRepo) ListByCandidateAndResume(ctx context.Context, candidateID, resumeID uuid.UUID) ([]*skill.CandidateSkill, error) {
	const q = `SELECT id, candidate_id, skill_id, resume_id, score, confidence, evidence_source, raw_evidence, created_at
		FROM candidate_skills WHERE candidate_id=$1 AND resume_id=$2 ORDER BY score DESC`
	rows, err := r.pool.Query(ctx, q, candidateID, resumeID)
	if err != nil {
		return nil, fmt.Errorf("candidate_skill_repo: list_by_candidate_resume: %w", err)
	}
	defer rows.Close()
	return scanCandidateSkills(rows)
}

func (r *CandidateSkillRepo) DeleteByResumeID(ctx context.Context, resumeID uuid.UUID) error {
	const q = `DELETE FROM candidate_skills WHERE resume_id=$1`
	_, err := r.pool.Exec(ctx, q, resumeID)
	if err != nil {
		return fmt.Errorf("candidate_skill_repo: delete_by_resume: %w", err)
	}
	return nil
}

func scanCandidateSkills(rows pgx.Rows) ([]*skill.CandidateSkill, error) {
	var out []*skill.CandidateSkill
	for rows.Next() {
		var cs skill.CandidateSkill
		if err := rows.Scan(&cs.ID, &cs.CandidateID, &cs.SkillID, &cs.ResumeID,
			&cs.Score, &cs.Confidence, &cs.EvidenceSource, &cs.RawEvidence, &cs.CreatedAt); err != nil {
			return nil, fmt.Errorf("candidate_skill_repo: scan: %w", err)
		}
		out = append(out, &cs)
	}
	return out, rows.Err()
}
