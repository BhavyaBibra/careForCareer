package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"careergps/internal/domain/roadmap"
	"careergps/pkg/apperrors"
)

type RoadmapRepo struct {
	pool *pgxpool.Pool
}

func NewRoadmapRepo(pool *pgxpool.Pool) *RoadmapRepo {
	return &RoadmapRepo{pool: pool}
}

func (r *RoadmapRepo) Create(ctx context.Context, rm *roadmap.Roadmap) error {
	dayPlansJSON, err := json.Marshal(rm.DayPlans)
	if err != nil {
		return fmt.Errorf("roadmap: marshal day_plans: %w", err)
	}
	_, err = r.pool.Exec(ctx, `
		INSERT INTO roadmaps (id, candidate_id, assessment_id, interview_date, plan_start_date,
			daily_tasks, narrative, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		rm.ID, rm.CandidateID, rm.AssessmentID,
		rm.InterviewDate, rm.PlanStartDate,
		dayPlansJSON, rm.Narrative,
		rm.CreatedAt, rm.UpdatedAt,
	)
	return err
}

func (r *RoadmapRepo) GetByAssessmentID(ctx context.Context, assessmentID uuid.UUID) (*roadmap.Roadmap, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, candidate_id, assessment_id, interview_date, plan_start_date,
			daily_tasks, narrative, created_at, updated_at
		FROM roadmaps WHERE assessment_id=$1 ORDER BY created_at DESC LIMIT 1`, assessmentID)
	return scanRoadmap(row)
}

func (r *RoadmapRepo) GetByID(ctx context.Context, id uuid.UUID) (*roadmap.Roadmap, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, candidate_id, assessment_id, interview_date, plan_start_date,
			daily_tasks, narrative, created_at, updated_at
		FROM roadmaps WHERE id=$1`, id)
	return scanRoadmap(row)
}

func (r *RoadmapRepo) Update(ctx context.Context, rm *roadmap.Roadmap) error {
	dayPlansJSON, err := json.Marshal(rm.DayPlans)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, `
		UPDATE roadmaps SET daily_tasks=$1, narrative=$2, updated_at=$3 WHERE id=$4`,
		dayPlansJSON, rm.Narrative, rm.UpdatedAt, rm.ID,
	)
	return err
}

func scanRoadmap(row pgx.Row) (*roadmap.Roadmap, error) {
	var rm roadmap.Roadmap
	var dayPlansJSON []byte
	err := row.Scan(
		&rm.ID, &rm.CandidateID, &rm.AssessmentID,
		&rm.InterviewDate, &rm.PlanStartDate,
		&dayPlansJSON, &rm.Narrative,
		&rm.CreatedAt, &rm.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("roadmap: scan: %w", err)
	}
	if err := json.Unmarshal(dayPlansJSON, &rm.DayPlans); err != nil {
		return nil, fmt.Errorf("roadmap: unmarshal day_plans: %w", err)
	}
	return &rm, nil
}
