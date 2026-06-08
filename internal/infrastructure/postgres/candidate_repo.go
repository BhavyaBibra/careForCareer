package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"careergps/internal/domain/candidate"
	"careergps/pkg/apperrors"
)

type CandidateRepo struct {
	pool *pgxpool.Pool
}

func NewCandidateRepo(pool *pgxpool.Pool) *CandidateRepo {
	return &CandidateRepo{pool: pool}
}

func (r *CandidateRepo) Create(ctx context.Context, c *candidate.Candidate) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO candidates (id, user_id, years_experience, inferred_tier, tier_explanation,
			current_company, current_comp_inr, target_comp_inr, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		c.ID, c.UserID, c.YearsExperience, int(c.InferredTier), c.TierExplanation,
		c.CurrentCompany, int64(c.CurrentComp), int64(c.TargetComp),
		c.CreatedAt, c.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("candidate: create: %w", err)
	}
	return nil
}

func (r *CandidateRepo) GetByID(ctx context.Context, id uuid.UUID) (*candidate.Candidate, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, years_experience, inferred_tier, tier_explanation,
			current_company, current_comp_inr, target_comp_inr, created_at, updated_at
		FROM candidates WHERE id = $1`, id)
	return scanCandidate(row)
}

func (r *CandidateRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (*candidate.Candidate, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, years_experience, inferred_tier, tier_explanation,
			current_company, current_comp_inr, target_comp_inr, created_at, updated_at
		FROM candidates WHERE user_id = $1`, userID)
	return scanCandidate(row)
}

func (r *CandidateRepo) Update(ctx context.Context, c *candidate.Candidate) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE candidates SET years_experience=$1, inferred_tier=$2, tier_explanation=$3,
			current_company=$4, current_comp_inr=$5, target_comp_inr=$6, updated_at=$7
		WHERE id=$8`,
		c.YearsExperience, int(c.InferredTier), c.TierExplanation,
		c.CurrentCompany, int64(c.CurrentComp), int64(c.TargetComp),
		c.UpdatedAt, c.ID,
	)
	return err
}

func (r *CandidateRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM candidates WHERE id=$1`, id)
	return err
}

func scanCandidate(row pgx.Row) (*candidate.Candidate, error) {
	var c candidate.Candidate
	var tier int
	var currentComp, targetComp int64
	err := row.Scan(
		&c.ID, &c.UserID, &c.YearsExperience, &tier, &c.TierExplanation,
		&c.CurrentCompany, &currentComp, &targetComp, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("candidate: scan: %w", err)
	}
	c.InferredTier = candidate.ExperienceTier(tier)
	c.CurrentComp = candidate.CompensationINR(currentComp)
	c.TargetComp = candidate.CompensationINR(targetComp)
	return &c, nil
}
