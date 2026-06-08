package readiness

import (
	"errors"
	"math"
	"time"

	"github.com/google/uuid"

	"careergps/internal/domain/candidate"
)

// TierWeights defines the scoring weight distribution for a given experience tier.
// Invariant: all weights must sum to 1.0 — validated at startup via ValidateWeightTable.
// Changing weights requires incrementing EngineVersion in config.
type TierWeights struct {
	SkillMatch      float64 `yaml:"skill_match"`
	DSASignal       float64 `yaml:"dsa_signal"`
	SystemDesign    float64 `yaml:"system_design"`
	ArchDepth       float64 `yaml:"arch_depth"`
	DomainRelevance float64 `yaml:"domain_relevance"`
	ExperienceMatch float64 `yaml:"experience_match"`
}

func (w TierWeights) Sum() float64 {
	return w.SkillMatch + w.DSASignal + w.SystemDesign +
		w.ArchDepth + w.DomainRelevance + w.ExperienceMatch
}

// Valid returns true when weights sum to 1.0 (within float64 tolerance).
func (w TierWeights) Valid() bool {
	s := w.Sum()
	return math.Abs(s-1.0) < 0.001
}

// DefaultWeightTable is used when config file is absent or incomplete.
// Do not modify — change scoring_weights.yaml and increment engine_version instead.
var DefaultWeightTable = map[candidate.ExperienceTier]TierWeights{
	candidate.TierFreshGrad: {SkillMatch: 0.40, DSASignal: 0.30, SystemDesign: 0.10, ArchDepth: 0.00, DomainRelevance: 0.10, ExperienceMatch: 0.10},
	candidate.TierJunior:    {SkillMatch: 0.35, DSASignal: 0.25, SystemDesign: 0.15, ArchDepth: 0.05, DomainRelevance: 0.10, ExperienceMatch: 0.10},
	candidate.TierMidLevel:  {SkillMatch: 0.30, DSASignal: 0.15, SystemDesign: 0.25, ArchDepth: 0.15, DomainRelevance: 0.10, ExperienceMatch: 0.05},
	candidate.TierSenior:    {SkillMatch: 0.20, DSASignal: 0.05, SystemDesign: 0.30, ArchDepth: 0.35, DomainRelevance: 0.05, ExperienceMatch: 0.05},
	candidate.TierStaff:     {SkillMatch: 0.15, DSASignal: 0.05, SystemDesign: 0.25, ArchDepth: 0.40, DomainRelevance: 0.10, ExperienceMatch: 0.05},
}

// ValidateWeightTable validates all tier weights at startup. Fail fast.
func ValidateWeightTable(table map[candidate.ExperienceTier]TierWeights) error {
	for tier, w := range table {
		if !w.Valid() {
			return errors.New("scoring: tier weights do not sum to 1.0")
		}
		// Tier 0 (FreshGrad): arch_depth must be 0 — not expected, excluded entirely.
		if tier == candidate.TierFreshGrad && w.ArchDepth != 0.0 {
			return errors.New("scoring: tier 0 (fresh_grad) arch_depth must be 0.0")
		}
	}
	return nil
}

// ComponentScores holds the [0, 100] score for each scoring component.
// All values are deterministic — computed from CandidateSkills and JDSkills.
type ComponentScores struct {
	SkillMatch      float64 `json:"skill_match"`
	DSASignal       float64 `json:"dsa_signal"`
	SystemDesign    float64 `json:"system_design"`
	ArchDepth       float64 `json:"arch_depth"`
	DomainRelevance float64 `json:"domain_relevance"`
	ExperienceMatch float64 `json:"experience_match"`
}

// Score computes the composite readiness score.
// Deterministic — same inputs always produce same output.
// Result is in [0, 100].
func Score(components ComponentScores, weights TierWeights) float64 {
	raw := components.SkillMatch*weights.SkillMatch +
		components.DSASignal*weights.DSASignal +
		components.SystemDesign*weights.SystemDesign +
		components.ArchDepth*weights.ArchDepth +
		components.DomainRelevance*weights.DomainRelevance +
		components.ExperienceMatch*weights.ExperienceMatch

	// Clamp to [0, 100]
	if raw < 0 {
		return 0
	}
	if raw > 100 {
		return 100
	}
	return raw
}

// ReadinessAssessment is the immutable output of the scoring engine.
// Invariant: once persisted, never mutated — new assessment = new record.
// EngineVersion ties this record to a specific weight table snapshot.
// Historical assessments remain queryable with their original engine version.
type ReadinessAssessment struct {
	ID             uuid.UUID
	CandidateID    uuid.UUID
	JDID           uuid.UUID
	GapAnalysisID  uuid.UUID
	Tier           candidate.ExperienceTier
	EngineVersion  string          // e.g. "v1.0.0"
	CompositeScore float64         // [0, 100]
	Components     ComponentScores // snapshot of component scores
	WeightsUsed    TierWeights     // snapshot of weights at scoring time — for audit
	InputSnapshot  string          // JSON snapshot of full inputs — for auditability
	CreatedAt      time.Time
}

func (ra *ReadinessAssessment) Validate() error {
	if ra.CompositeScore < 0 || ra.CompositeScore > 100 {
		return errors.New("readiness: composite score out of [0, 100]")
	}
	if !ra.WeightsUsed.Valid() {
		return errors.New("readiness: weights do not sum to 1.0")
	}
	if ra.EngineVersion == "" {
		return errors.New("readiness: EngineVersion is required")
	}
	return nil
}

// ComponentContribution returns the contribution of each component to the final score.
func (ra *ReadinessAssessment) ComponentContribution() map[string]float64 {
	return map[string]float64{
		"skill_match":      ra.Components.SkillMatch * ra.WeightsUsed.SkillMatch,
		"dsa_signal":       ra.Components.DSASignal * ra.WeightsUsed.DSASignal,
		"system_design":    ra.Components.SystemDesign * ra.WeightsUsed.SystemDesign,
		"arch_depth":       ra.Components.ArchDepth * ra.WeightsUsed.ArchDepth,
		"domain_relevance": ra.Components.DomainRelevance * ra.WeightsUsed.DomainRelevance,
		"experience_match": ra.Components.ExperienceMatch * ra.WeightsUsed.ExperienceMatch,
	}
}
