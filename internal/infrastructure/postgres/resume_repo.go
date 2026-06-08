package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"careergps/internal/domain/resume"
	"careergps/pkg/apperrors"
)

type ResumeRepo struct {
	pool *pgxpool.Pool
}

func NewResumeRepo(pool *pgxpool.Pool) *ResumeRepo {
	return &ResumeRepo{pool: pool}
}

func (r *ResumeRepo) Create(ctx context.Context, res *resume.Resume) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO resumes (id, candidate_id, source_type, storage_key, github_url,
			raw_text, extraction_status, extraction_error, version, parse_attempts, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		res.ID, res.CandidateID, string(res.SourceType), res.StorageKey, res.GitHubURL,
		res.RawText, string(res.ExtractionStatus), res.ExtractionError,
		res.Version, res.ParseAttempts, res.CreatedAt, res.UpdatedAt,
	)
	return err
}

func (r *ResumeRepo) GetByID(ctx context.Context, id uuid.UUID) (*resume.Resume, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, candidate_id, source_type, storage_key, github_url,
			raw_text, extraction_status, extraction_error, version, parse_attempts, created_at, updated_at
		FROM resumes WHERE id=$1`, id)
	return scanResume(row)
}

func (r *ResumeRepo) GetLatestByCandidateID(ctx context.Context, candidateID uuid.UUID) (*resume.Resume, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, candidate_id, source_type, storage_key, github_url,
			raw_text, extraction_status, extraction_error, version, parse_attempts, created_at, updated_at
		FROM resumes WHERE candidate_id=$1 ORDER BY version DESC LIMIT 1`, candidateID)
	return scanResume(row)
}

func (r *ResumeRepo) Update(ctx context.Context, res *resume.Resume) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE resumes SET raw_text=$1, extraction_status=$2, extraction_error=$3,
			parse_attempts=$4, updated_at=$5
		WHERE id=$6`,
		res.RawText, string(res.ExtractionStatus), res.ExtractionError,
		res.ParseAttempts, res.UpdatedAt, res.ID,
	)
	return err
}

func (r *ResumeRepo) NextVersion(ctx context.Context, candidateID uuid.UUID) (int, error) {
	var max *int
	err := r.pool.QueryRow(ctx, `SELECT MAX(version) FROM resumes WHERE candidate_id=$1`, candidateID).Scan(&max)
	if err != nil {
		return 0, err
	}
	if max == nil {
		return 1, nil
	}
	return *max + 1, nil
}

func scanResume(row pgx.Row) (*resume.Resume, error) {
	var res resume.Resume
	var sourceType, status string
	err := row.Scan(
		&res.ID, &res.CandidateID, &sourceType, &res.StorageKey, &res.GitHubURL,
		&res.RawText, &status, &res.ExtractionError,
		&res.Version, &res.ParseAttempts, &res.CreatedAt, &res.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("resume: scan: %w", err)
	}
	res.SourceType = resume.SourceType(sourceType)
	res.ExtractionStatus = resume.ExtractionStatus(status)
	return &res, nil
}
