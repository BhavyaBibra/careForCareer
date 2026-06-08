package gap

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type FitLevel string

const (
	FitStrong  FitLevel = "strong"  // readiness >= 75
	FitStretch FitLevel = "stretch" // readiness 50–74
	FitNotYet  FitLevel = "not_yet" // readiness < 50
)

func FitFromScore(score float64) FitLevel {
	switch {
	case score >= 75:
		return FitStrong
	case score >= 50:
		return FitStretch
	default:
		return FitNotYet
	}
}

// SkillGap is the per-skill deficit for a candidate against a specific JD.
// Gap = JDRequiredScore - CandidateScore, clamped to min 0 (surplus is 0, not negative).
// WeightedPriority drives the recommendation engine sort order.
type SkillGap struct {
	SkillID          uuid.UUID `json:"skill_id"`
	SkillName        string    `json:"skill_name"`
	CandidateScore   int       `json:"candidate_score"` // 0 if skill absent from profile
	JDRequiredScore  int       `json:"jd_required_score"`
	Gap              int       `json:"gap"`              // clamped to [0, 9]
	IsRequired       bool      `json:"is_required"`
	WeightedPriority float64   `json:"weighted_priority"`
}

// ComputeGap calculates the gap and clamps to zero minimum.
func ComputeGap(candidateScore, jdRequiredScore int) int {
	g := jdRequiredScore - candidateScore
	if g < 0 {
		return 0
	}
	return g
}

// GapAnalysis aggregates all per-skill gaps for a candidate × JD pair.
// AggregateGap is the demand-weighted average gap across required skills only.
// Confidence: fraction of required JD skills that had a corresponding candidate skill signal.
// Invariant: once persisted, GapAnalysis is immutable.
type GapAnalysis struct {
	ID           uuid.UUID  `json:"id"`
	CandidateID  uuid.UUID  `json:"candidate_id"`
	JDID         uuid.UUID  `json:"jd_id"`
	AssessmentID *uuid.UUID `json:"assessment_id,omitempty"` // set after readiness computed
	Gaps         []SkillGap `json:"gaps"`
	AggregateGap float64    `json:"aggregate_gap"`
	Confidence   float64    `json:"confidence"` // [0, 1]
	FitLevel     FitLevel   `json:"fit_level"`
	CreatedAt    time.Time  `json:"created_at"`
}

func (ga *GapAnalysis) Validate() error {
	if ga.CandidateID == uuid.Nil || ga.JDID == uuid.Nil {
		return errors.New("gap_analysis: CandidateID and JDID are required")
	}
	if ga.Confidence < 0 || ga.Confidence > 1 {
		return errors.New("gap_analysis: Confidence must be in [0, 1]")
	}
	return nil
}

// CriticalGaps returns skills with gap >= threshold (default: 3 points).
func (ga *GapAnalysis) CriticalGaps(threshold int) []SkillGap {
	var out []SkillGap
	for _, g := range ga.Gaps {
		if g.Gap >= threshold {
			out = append(out, g)
		}
	}
	return out
}

// OnTrackSkills returns skills where the candidate meets or exceeds the requirement.
func (ga *GapAnalysis) OnTrackSkills() []SkillGap {
	var out []SkillGap
	for _, g := range ga.Gaps {
		if g.Gap == 0 {
			out = append(out, g)
		}
	}
	return out
}
