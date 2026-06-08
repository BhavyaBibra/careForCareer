package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"careergps/internal/domain/company"
)

type CompanyRepo struct {
	pool *pgxpool.Pool
}

func NewCompanyRepo(pool *pgxpool.Pool) *CompanyRepo {
	return &CompanyRepo{pool: pool}
}

func (r *CompanyRepo) Create(ctx context.Context, c *company.Company) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	now := time.Now().UTC()
	c.CreatedAt = now
	c.UpdatedAt = now
	const q = `INSERT INTO companies (id, name, tier, india_bar_notes, website, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`
	_, err := r.pool.Exec(ctx, q,
		c.ID, c.Name, string(c.Tier), c.IndiaBarNotes, c.Website, c.CreatedAt, c.UpdatedAt)
	if err != nil {
		return fmt.Errorf("company_repo: create: %w", err)
	}
	return nil
}

func (r *CompanyRepo) GetByID(ctx context.Context, id uuid.UUID) (*company.Company, error) {
	const q = `SELECT id, name, tier, india_bar_notes, website, created_at, updated_at
		FROM companies WHERE id=$1`
	row := r.pool.QueryRow(ctx, q, id)
	return scanCompany(row)
}

func (r *CompanyRepo) GetByName(ctx context.Context, name string) (*company.Company, error) {
	const q = `SELECT id, name, tier, india_bar_notes, website, created_at, updated_at
		FROM companies WHERE lower(name)=lower($1)`
	row := r.pool.QueryRow(ctx, q, name)
	return scanCompany(row)
}

func (r *CompanyRepo) List(ctx context.Context, tier *company.Tier, limit int, cursor string) ([]*company.Company, string, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	var rows pgx.Rows
	var err error

	if tier != nil {
		if cursor == "" {
			const q = `SELECT id, name, tier, india_bar_notes, website, created_at, updated_at
				FROM companies WHERE tier=$1 ORDER BY name LIMIT $2`
			rows, err = r.pool.Query(ctx, q, string(*tier), limit+1)
		} else {
			const q = `SELECT id, name, tier, india_bar_notes, website, created_at, updated_at
				FROM companies WHERE tier=$1 AND name > $2 ORDER BY name LIMIT $3`
			rows, err = r.pool.Query(ctx, q, string(*tier), cursor, limit+1)
		}
	} else {
		if cursor == "" {
			const q = `SELECT id, name, tier, india_bar_notes, website, created_at, updated_at
				FROM companies ORDER BY name LIMIT $1`
			rows, err = r.pool.Query(ctx, q, limit+1)
		} else {
			const q = `SELECT id, name, tier, india_bar_notes, website, created_at, updated_at
				FROM companies WHERE name > $1 ORDER BY name LIMIT $2`
			rows, err = r.pool.Query(ctx, q, cursor, limit+1)
		}
	}
	if err != nil {
		return nil, "", fmt.Errorf("company_repo: list: %w", err)
	}
	defer rows.Close()

	var results []*company.Company
	for rows.Next() {
		c, scanErr := scanCompany(rows)
		if scanErr != nil {
			return nil, "", scanErr
		}
		results = append(results, c)
	}
	if err := rows.Err(); err != nil {
		return nil, "", fmt.Errorf("company_repo: list rows: %w", err)
	}

	var nextCursor string
	if len(results) > limit {
		results = results[:limit]
		nextCursor = results[len(results)-1].Name
	}
	return results, nextCursor, nil
}

func (r *CompanyRepo) Update(ctx context.Context, c *company.Company) error {
	c.UpdatedAt = time.Now().UTC()
	const q = `UPDATE companies SET name=$2, tier=$3, india_bar_notes=$4, website=$5, updated_at=$6 WHERE id=$1`
	_, err := r.pool.Exec(ctx, q, c.ID, c.Name, string(c.Tier), c.IndiaBarNotes, c.Website, c.UpdatedAt)
	if err != nil {
		return fmt.Errorf("company_repo: update: %w", err)
	}
	return nil
}

func scanCompany(row interface {
	Scan(dest ...any) error
}) (*company.Company, error) {
	var c company.Company
	var tier string
	err := row.Scan(&c.ID, &c.Name, &tier, &c.IndiaBarNotes, &c.Website, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("company_repo: not found")
	}
	if err != nil {
		return nil, fmt.Errorf("company_repo: scan: %w", err)
	}
	c.Tier = company.Tier(tier)
	return &c, nil
}

// CompanyPatternRepo implements company.PatternRepository.
type CompanyPatternRepo struct {
	pool *pgxpool.Pool
}

func NewCompanyPatternRepo(pool *pgxpool.Pool) *CompanyPatternRepo {
	return &CompanyPatternRepo{pool: pool}
}

func (r *CompanyPatternRepo) Upsert(ctx context.Context, p *company.CompanyPattern) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	now := time.Now().UTC()
	p.CreatedAt = now
	p.UpdatedAt = now

	roundsJSON, err := json.Marshal(p.InterviewRounds)
	if err != nil {
		return fmt.Errorf("company_pattern_repo: marshal rounds: %w", err)
	}
	focusJSON, err := json.Marshal(p.FocusAreas)
	if err != nil {
		return fmt.Errorf("company_pattern_repo: marshal focus: %w", err)
	}
	rejectionJSON, err := json.Marshal(p.TypicalRejectionReasons)
	if err != nil {
		return fmt.Errorf("company_pattern_repo: marshal rejection: %w", err)
	}

	const q = `INSERT INTO company_patterns
		(id, company_id, interview_rounds, focus_areas, typical_rejection_reasons,
		 dsa_difficulty, notes, updated_by, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		ON CONFLICT (company_id) DO UPDATE SET
		interview_rounds=$3, focus_areas=$4, typical_rejection_reasons=$5,
		dsa_difficulty=$6, notes=$7, updated_by=$8, updated_at=$10`
	_, err = r.pool.Exec(ctx, q,
		p.ID, p.CompanyID,
		roundsJSON, focusJSON, rejectionJSON,
		p.DSADifficulty, p.Notes, p.UpdatedBy,
		p.CreatedAt, p.UpdatedAt)
	if err != nil {
		return fmt.Errorf("company_pattern_repo: upsert: %w", err)
	}
	return nil
}

func (r *CompanyPatternRepo) GetByCompanyID(ctx context.Context, companyID uuid.UUID) (*company.CompanyPattern, error) {
	const q = `SELECT id, company_id, interview_rounds, focus_areas, typical_rejection_reasons,
		dsa_difficulty, notes, updated_by, created_at, updated_at
		FROM company_patterns WHERE company_id=$1`
	row := r.pool.QueryRow(ctx, q, companyID)

	var p company.CompanyPattern
	var roundsJSON, focusJSON, rejectionJSON []byte
	err := row.Scan(
		&p.ID, &p.CompanyID,
		&roundsJSON, &focusJSON, &rejectionJSON,
		&p.DSADifficulty, &p.Notes, &p.UpdatedBy,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("company_pattern_repo: not found")
	}
	if err != nil {
		return nil, fmt.Errorf("company_pattern_repo: scan: %w", err)
	}
	if err := json.Unmarshal(roundsJSON, &p.InterviewRounds); err != nil {
		return nil, fmt.Errorf("company_pattern_repo: unmarshal rounds: %w", err)
	}
	if err := json.Unmarshal(focusJSON, &p.FocusAreas); err != nil {
		return nil, fmt.Errorf("company_pattern_repo: unmarshal focus: %w", err)
	}
	if err := json.Unmarshal(rejectionJSON, &p.TypicalRejectionReasons); err != nil {
		return nil, fmt.Errorf("company_pattern_repo: unmarshal rejection: %w", err)
	}
	return &p, nil
}
