package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"careergps/internal/domain/candidate"
	"careergps/internal/domain/gap"
	"careergps/internal/domain/readiness"
	"careergps/pkg/apperrors"
)

// GapAnalysisRepo handles persistence for gap.GapAnalysis.
type GapAnalysisRepo struct {
	pool *pgxpool.Pool
}

func NewGapAnalysisRepo(pool *pgxpool.Pool) *GapAnalysisRepo {
	return &GapAnalysisRepo{pool: pool}
}

func (r *GapAnalysisRepo) Create(ctx context.Context, ga *gap.GapAnalysis) error {
	gapsJSON, err := json.Marshal(ga.Gaps)
	if err != nil {
		return fmt.Errorf("gap_analysis: marshal gaps: %w", err)
	}
	_, err = r.pool.Exec(ctx, `
		INSERT INTO gap_analyses (id, candidate_id, jd_id, gaps_json, aggregate_gap, confidence, fit_level, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		ga.ID, ga.CandidateID, ga.JDID, gapsJSON, ga.AggregateGap, ga.Confidence, string(ga.FitLevel), ga.CreatedAt,
	)
	return err
}

func (r *GapAnalysisRepo) GetByID(ctx context.Context, id uuid.UUID) (*gap.GapAnalysis, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, candidate_id, jd_id, gaps_json, aggregate_gap, confidence, fit_level, created_at
		FROM gap_analyses WHERE id=$1`, id)
	return scanGapAnalysis(row)
}

func (r *GapAnalysisRepo) GetLatestByCandidateAndJD(ctx context.Context, candidateID, jdID uuid.UUID) (*gap.GapAnalysis, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, candidate_id, jd_id, gaps_json, aggregate_gap, confidence, fit_level, created_at
		FROM gap_analyses WHERE candidate_id=$1 AND jd_id=$2 ORDER BY created_at DESC LIMIT 1`,
		candidateID, jdID)
	return scanGapAnalysis(row)
}

func scanGapAnalysis(row pgx.Row) (*gap.GapAnalysis, error) {
	var ga gap.GapAnalysis
	var gapsJSON []byte
	var fitLevel string
	err := row.Scan(&ga.ID, &ga.CandidateID, &ga.JDID, &gapsJSON, &ga.AggregateGap, &ga.Confidence, &fitLevel, &ga.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("gap_analysis: scan: %w", err)
	}
	if err := json.Unmarshal(gapsJSON, &ga.Gaps); err != nil {
		return nil, fmt.Errorf("gap_analysis: unmarshal gaps: %w", err)
	}
	ga.FitLevel = gap.FitLevel(fitLevel)
	return &ga, nil
}

// ReadinessRepo handles persistence for readiness.ReadinessAssessment.
type ReadinessRepo struct {
	pool *pgxpool.Pool
}

func NewReadinessRepo(pool *pgxpool.Pool) *ReadinessRepo {
	return &ReadinessRepo{pool: pool}
}

func (r *ReadinessRepo) Create(ctx context.Context, ra *readiness.ReadinessAssessment) error {
	componentsJSON, err := json.Marshal(ra.Components)
	if err != nil {
		return err
	}
	weightsJSON, err := json.Marshal(ra.WeightsUsed)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, `
		INSERT INTO readiness_assessments
		(id, candidate_id, jd_id, gap_analysis_id, tier, engine_version, composite_score,
		 components_json, weights_json, input_snapshot, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		ra.ID, ra.CandidateID, ra.JDID, ra.GapAnalysisID,
		int(ra.Tier), ra.EngineVersion, ra.CompositeScore,
		componentsJSON, weightsJSON, ra.InputSnapshot, ra.CreatedAt,
	)
	return err
}

func (r *ReadinessRepo) GetByID(ctx context.Context, id uuid.UUID) (*readiness.ReadinessAssessment, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, candidate_id, jd_id, gap_analysis_id, tier, engine_version, composite_score,
			components_json, weights_json, input_snapshot, created_at
		FROM readiness_assessments WHERE id=$1`, id)
	return scanAssessment(row)
}

func (r *ReadinessRepo) GetLatestByCandidateAndJD(ctx context.Context, candidateID, jdID uuid.UUID) (*readiness.ReadinessAssessment, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, candidate_id, jd_id, gap_analysis_id, tier, engine_version, composite_score,
			components_json, weights_json, input_snapshot, created_at
		FROM readiness_assessments WHERE candidate_id=$1 AND jd_id=$2 ORDER BY created_at DESC LIMIT 1`,
		candidateID, jdID)
	return scanAssessment(row)
}

func (r *ReadinessRepo) ListByCandidateID(ctx context.Context, candidateID uuid.UUID, limit int, cursor string) ([]*readiness.ReadinessAssessment, string, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, candidate_id, jd_id, gap_analysis_id, tier, engine_version, composite_score,
			components_json, weights_json, input_snapshot, created_at
		FROM readiness_assessments WHERE candidate_id=$1 ORDER BY created_at DESC LIMIT $2`,
		candidateID, limit)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()
	var out []*readiness.ReadinessAssessment
	for rows.Next() {
		ra, err := scanAssessment(rows)
		if err != nil {
			return nil, "", err
		}
		out = append(out, ra)
	}
	return out, "", nil
}

func scanAssessment(row interface {
	Scan(dest ...interface{}) error
}) (*readiness.ReadinessAssessment, error) {
	var ra readiness.ReadinessAssessment
	var tierInt int
	var componentsJSON, weightsJSON []byte
	err := row.Scan(
		&ra.ID, &ra.CandidateID, &ra.JDID, &ra.GapAnalysisID,
		&tierInt, &ra.EngineVersion, &ra.CompositeScore,
		&componentsJSON, &weightsJSON, &ra.InputSnapshot, &ra.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("readiness: scan: %w", err)
	}
	ra.Tier = candidate.ExperienceTier(tierInt)
	if err := json.Unmarshal(componentsJSON, &ra.Components); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(weightsJSON, &ra.WeightsUsed); err != nil {
		return nil, err
	}
	return &ra, nil
}
