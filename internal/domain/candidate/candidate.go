package candidate

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ExperienceTier is inferred from YOE — never set by the user directly.
type ExperienceTier int

const (
	TierFreshGrad ExperienceTier = 0 // 0 YOE
	TierJunior    ExperienceTier = 1 // 1–3 YOE
	TierMidLevel  ExperienceTier = 2 // 3–6 YOE
	TierSenior    ExperienceTier = 3 // 6–10 YOE
	TierStaff     ExperienceTier = 4 // 10+ YOE
)

// TierLabel returns a human-readable label for the tier.
func (t ExperienceTier) TierLabel() string {
	switch t {
	case TierFreshGrad:
		return "Fresh Grad"
	case TierJunior:
		return "Junior (SDE-1)"
	case TierMidLevel:
		return "Mid-level (SDE-2)"
	case TierSenior:
		return "Senior (SDE-3)"
	case TierStaff:
		return "Staff+"
	default:
		return "Unknown"
	}
}

func (t ExperienceTier) Valid() bool { return t >= 0 && t <= 4 }

// tierBoundary maps minimum YOE to tier. Ordered descending — first match wins.
var tierBoundaries = []struct {
	MinYOE int
	Tier   ExperienceTier
}{
	{10, TierStaff},
	{6, TierSenior},
	{3, TierMidLevel},
	{1, TierJunior},
	{0, TierFreshGrad},
}

// InferTier derives the experience tier from years of experience.
// Deterministic — no LLM involved.
func InferTier(yoe int) ExperienceTier {
	for _, b := range tierBoundaries {
		if yoe >= b.MinYOE {
			return b.Tier
		}
	}
	return TierFreshGrad
}

// CompensationINR is annual compensation in Indian Rupees. Zero means not provided.
type CompensationINR int64

// Candidate is the root aggregate for a user's career profile.
// Invariant: UserID is always set and unique.
// Invariant: InferredTier is computed from YearsExperience, never stored as user input.
// Invariant: TargetCompanyIDs may be empty — system suggests companies when absent.
type Candidate struct {
	ID               uuid.UUID
	UserID           uuid.UUID // FK to users; enforces multi-tenancy
	YearsExperience  int       // Invariant: >= 0
	InferredTier     ExperienceTier
	TierExplanation  string // e.g. "4 YOE at Flipkart → Mid-level (SDE-2)"
	CurrentCompany   string
	CurrentComp      CompensationINR
	TargetComp       CompensationINR
	TargetCompanyIDs []uuid.UUID
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// New creates and validates a Candidate, computing the tier automatically.
func New(userID uuid.UUID, yoe int, currentCompany string, currentComp, targetComp CompensationINR, targetCompanyIDs []uuid.UUID) (*Candidate, error) {
	if userID == uuid.Nil {
		return nil, errors.New("candidate: UserID is required")
	}
	if yoe < 0 {
		return nil, errors.New("candidate: YearsExperience cannot be negative")
	}
	tier := InferTier(yoe)
	explanation := buildTierExplanation(yoe, tier, currentCompany)
	now := time.Now().UTC()
	return &Candidate{
		ID:               uuid.New(),
		UserID:           userID,
		YearsExperience:  yoe,
		InferredTier:     tier,
		TierExplanation:  explanation,
		CurrentCompany:   currentCompany,
		CurrentComp:      currentComp,
		TargetComp:       targetComp,
		TargetCompanyIDs: targetCompanyIDs,
		CreatedAt:        now,
		UpdatedAt:        now,
	}, nil
}

func (c *Candidate) Validate() error {
	if c.UserID == uuid.Nil {
		return errors.New("candidate: UserID is required")
	}
	if c.YearsExperience < 0 {
		return errors.New("candidate: YearsExperience cannot be negative")
	}
	if !c.InferredTier.Valid() {
		return errors.New("candidate: InferredTier out of range [0,4]")
	}
	return nil
}

// UpdateProfile re-computes the tier whenever profile fields change.
func (c *Candidate) UpdateProfile(yoe int, currentCompany string, currentComp, targetComp CompensationINR, targetCompanyIDs []uuid.UUID) error {
	if yoe < 0 {
		return errors.New("candidate: YearsExperience cannot be negative")
	}
	c.YearsExperience = yoe
	c.InferredTier = InferTier(yoe)
	c.TierExplanation = buildTierExplanation(yoe, c.InferredTier, currentCompany)
	c.CurrentCompany = currentCompany
	c.CurrentComp = currentComp
	c.TargetComp = targetComp
	c.TargetCompanyIDs = targetCompanyIDs
	c.UpdatedAt = time.Now().UTC()
	return nil
}

func buildTierExplanation(yoe int, tier ExperienceTier, company string) string {
	base := ""
	if company != "" {
		base = company + ", "
	}
	return base + fmt.Sprintf("%d YOE → %s", yoe, tier.TierLabel())
}
