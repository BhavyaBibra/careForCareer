package skill

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Category string

const (
	CategoryDSA          Category = "dsa"
	CategoryBackend      Category = "backend"
	CategorySystemDesign Category = "system_design"
	CategoryArchitecture Category = "architecture"
	CategoryDomain       Category = "domain"
	CategoryLanguage     Category = "language"
	CategoryDevOps       Category = "devops"
)

// Skill is a canonical, normalised skill in the registry.
// Aliases are resolved to canonical Name at ingestion time.
// Invariant: Name is lowercase and trimmed.
// Invariant: Aliases are lowercase and trimmed.
type Skill struct {
	ID        uuid.UUID
	Name      string   // canonical lowercase, e.g. "go"
	Category  Category
	Aliases   []string // e.g. ["golang", "go lang", "go programming language"]
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (s *Skill) Validate() error {
	if s.Name == "" {
		return errors.New("skill: Name is required")
	}
	if s.Category == "" {
		return errors.New("skill: Category is required")
	}
	return nil
}

// CandidateSkill is the inferred skill score for a candidate from a specific resume version.
// Score is always in [1, 10]. Confidence is in [0.0, 1.0].
// Invariant: Score clamped before persistence via ClampScore.
// Invariant: low Confidence (<0.5) surfaces a UI warning — not silently accepted.
type CandidateSkill struct {
	ID             uuid.UUID
	CandidateID    uuid.UUID
	SkillID        uuid.UUID
	ResumeID       uuid.UUID
	Score          int     // [1, 10]
	Confidence     float64 // [0.0, 1.0]
	EvidenceSource string  // "resume_text" | "github_repo" | "manual" | "inferred"
	RawEvidence    string  // excerpt justifying the score
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// ClampScore enforces the [1, 10] invariant.
func ClampScore(s int) int {
	if s < 1 {
		return 1
	}
	if s > 10 {
		return 10
	}
	return s
}

func (cs *CandidateSkill) Validate() error {
	if cs.CandidateID == uuid.Nil || cs.SkillID == uuid.Nil {
		return errors.New("candidate_skill: CandidateID and SkillID are required")
	}
	if cs.Score < 1 || cs.Score > 10 {
		return errors.New("candidate_skill: Score must be in [1, 10]")
	}
	if cs.Confidence < 0 || cs.Confidence > 1 {
		return errors.New("candidate_skill: Confidence must be in [0.0, 1.0]")
	}
	if cs.EvidenceSource == "" {
		return errors.New("candidate_skill: EvidenceSource is required")
	}
	return nil
}

// LowConfidence returns true when the score should surface a UI warning.
const ConfidenceWarningThreshold = 0.5

func (cs *CandidateSkill) LowConfidence() bool {
	return cs.Confidence < ConfidenceWarningThreshold
}
