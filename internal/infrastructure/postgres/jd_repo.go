package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"careergps/internal/domain/jd"
)

type JDRepo struct {
	pool *pgxpool.Pool
}

func NewJDRepo(pool *pgxpool.Pool) *JDRepo {
	return &JDRepo{pool: pool}
}

func (r *JDRepo) Create(ctx context.Context, j *jd.JobDescription) error {
	if j.ID == uuid.Nil {
		j.ID = uuid.New()
	}
	now := time.Now().UTC()
	j.CreatedAt = now
	j.UpdatedAt = now
	const q = `INSERT INTO job_descriptions
		(id, candidate_id, company_id, raw_text, normalised_text,
		 seniority_signal, arch_expectation, extraction_status, extraction_error,
		 created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`
	_, err := r.pool.Exec(ctx, q,
		j.ID, j.CandidateID, j.CompanyID, j.RawText, j.NormalisedText,
		string(j.SenioritySignal), string(j.ArchExpectation),
		string(j.ExtractionStatus), j.ExtractionError,
		j.CreatedAt, j.UpdatedAt)
	if err != nil {
		return fmt.Errorf("jd_repo: create: %w", err)
	}
	return nil
}

func (r *JDRepo) GetByID(ctx context.Context, id uuid.UUID) (*jd.JobDescription, error) {
	const q = `SELECT id, candidate_id, company_id, raw_text, normalised_text,
		seniority_signal, arch_expectation, extraction_status, extraction_error,
		created_at, updated_at
		FROM job_descriptions WHERE id=$1`
	row := r.pool.QueryRow(ctx, q, id)
	return scanJD(row)
}

func (r *JDRepo) Update(ctx context.Context, j *jd.JobDescription) error {
	j.UpdatedAt = time.Now().UTC()
	const q = `UPDATE job_descriptions SET
		normalised_text=$2, seniority_signal=$3, arch_expectation=$4,
		extraction_status=$5, extraction_error=$6, company_id=$7, updated_at=$8
		WHERE id=$1`
	_, err := r.pool.Exec(ctx, q,
		j.ID, j.NormalisedText,
		string(j.SenioritySignal), string(j.ArchExpectation),
		string(j.ExtractionStatus), j.ExtractionError,
		j.CompanyID, j.UpdatedAt)
	if err != nil {
		return fmt.Errorf("jd_repo: update: %w", err)
	}
	return nil
}

func (r *JDRepo) ListByCandidateID(ctx context.Context, candidateID uuid.UUID, limit int, cursor string) ([]*jd.JobDescription, string, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	var rows pgx.Rows
	var err error

	if cursor == "" {
		const q = `SELECT id, candidate_id, company_id, raw_text, normalised_text,
			seniority_signal, arch_expectation, extraction_status, extraction_error,
			created_at, updated_at
			FROM job_descriptions WHERE candidate_id=$1 ORDER BY created_at DESC LIMIT $2`
		rows, err = r.pool.Query(ctx, q, candidateID, limit+1)
	} else {
		cursorTime, parseErr := time.Parse(time.RFC3339Nano, cursor)
		if parseErr != nil {
			return nil, "", fmt.Errorf("jd_repo: invalid cursor: %w", parseErr)
		}
		const q = `SELECT id, candidate_id, company_id, raw_text, normalised_text,
			seniority_signal, arch_expectation, extraction_status, extraction_error,
			created_at, updated_at
			FROM job_descriptions WHERE candidate_id=$1 AND created_at < $2 ORDER BY created_at DESC LIMIT $3`
		rows, err = r.pool.Query(ctx, q, candidateID, cursorTime, limit+1)
	}
	if err != nil {
		return nil, "", fmt.Errorf("jd_repo: list: %w", err)
	}
	defer rows.Close()

	var results []*jd.JobDescription
	for rows.Next() {
		j, scanErr := scanJD(rows)
		if scanErr != nil {
			return nil, "", scanErr
		}
		results = append(results, j)
	}
	if err := rows.Err(); err != nil {
		return nil, "", fmt.Errorf("jd_repo: list rows: %w", err)
	}

	var nextCursor string
	if len(results) > limit {
		results = results[:limit]
		nextCursor = results[len(results)-1].CreatedAt.Format(time.RFC3339Nano)
	}
	return results, nextCursor, nil
}

func scanJD(row interface {
	Scan(dest ...any) error
}) (*jd.JobDescription, error) {
	var j jd.JobDescription
	var seniority, arch, status string
	err := row.Scan(
		&j.ID, &j.CandidateID, &j.CompanyID,
		&j.RawText, &j.NormalisedText,
		&seniority, &arch, &status, &j.ExtractionError,
		&j.CreatedAt, &j.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("jd_repo: not found")
	}
	if err != nil {
		return nil, fmt.Errorf("jd_repo: scan: %w", err)
	}
	j.SenioritySignal = jd.SenioritySignal(seniority)
	j.ArchExpectation = jd.ArchExpectation(arch)
	j.ExtractionStatus = jd.ExtractionStatus(status)
	return &j, nil
}

// JDSkillRepo implements jd.JDSkillRepository.
type JDSkillRepo struct {
	pool *pgxpool.Pool
}

func NewJDSkillRepo(pool *pgxpool.Pool) *JDSkillRepo {
	return &JDSkillRepo{pool: pool}
}

func (r *JDSkillRepo) BulkCreate(ctx context.Context, skills []*jd.JDSkill) error {
	if len(skills) == 0 {
		return nil
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("jd_skill_repo: begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	const q = `INSERT INTO jd_skills
		(id, jd_id, skill_id, skill_name, is_required, min_required_score, weight)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (jd_id, skill_id) DO UPDATE
		SET skill_name=$4, is_required=$5, min_required_score=$6, weight=$7`

	for _, s := range skills {
		if s.ID == uuid.Nil {
			s.ID = uuid.New()
		}
		if _, err := tx.Exec(ctx, q,
			s.ID, s.JDID, s.SkillID, s.SkillName,
			s.IsRequired, s.MinRequiredScore, s.Weight); err != nil {
			return fmt.Errorf("jd_skill_repo: bulk_create: %w", err)
		}
	}
	return tx.Commit(ctx)
}

func (r *JDSkillRepo) ListByJDID(ctx context.Context, jdID uuid.UUID) ([]*jd.JDSkill, error) {
	const q = `SELECT id, jd_id, skill_id, skill_name, is_required, min_required_score, weight
		FROM jd_skills WHERE jd_id=$1 ORDER BY weight DESC`
	rows, err := r.pool.Query(ctx, q, jdID)
	if err != nil {
		return nil, fmt.Errorf("jd_skill_repo: list: %w", err)
	}
	defer rows.Close()

	var out []*jd.JDSkill
	for rows.Next() {
		var s jd.JDSkill
		if err := rows.Scan(&s.ID, &s.JDID, &s.SkillID, &s.SkillName,
			&s.IsRequired, &s.MinRequiredScore, &s.Weight); err != nil {
			return nil, fmt.Errorf("jd_skill_repo: scan: %w", err)
		}
		out = append(out, &s)
	}
	return out, rows.Err()
}

func (r *JDSkillRepo) DeleteByJDID(ctx context.Context, jdID uuid.UUID) error {
	const q = `DELETE FROM jd_skills WHERE jd_id=$1`
	_, err := r.pool.Exec(ctx, q, jdID)
	if err != nil {
		return fmt.Errorf("jd_skill_repo: delete: %w", err)
	}
	return nil
}
