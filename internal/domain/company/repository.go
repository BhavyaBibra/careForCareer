package company

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, c *Company) error
	GetByID(ctx context.Context, id uuid.UUID) (*Company, error)
	GetByName(ctx context.Context, name string) (*Company, error)
	List(ctx context.Context, tier *Tier, limit int, cursor string) ([]*Company, string, error)
	Update(ctx context.Context, c *Company) error
}

type PatternRepository interface {
	Upsert(ctx context.Context, p *CompanyPattern) error
	GetByCompanyID(ctx context.Context, companyID uuid.UUID) (*CompanyPattern, error)
}
