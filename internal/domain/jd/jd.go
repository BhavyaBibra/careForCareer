package jd

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type SenioritySignal string

const (
	SeniorityJunior  SenioritySignal = "junior"
	SeniorityMid     SenioritySignal = "mid"
	SenioritySenior  SenioritySignal = "senior"
	SeniorityStaff   SenioritySignal = "staff"
	SeniorityUnknown SenioritySignal = "unknown"
)

type ArchExpectation string

const (
	ArchNone     ArchExpectation = "none"     // IC contributor, no arch ownership expected
	ArchTeam     ArchExpectation = "team"     // team-scoped architecture
	ArchOrg      ArchExpectation = "org"      // cross-team architecture ownership
	ArchPlatform ArchExpectation = "platform" // platform/infra-level decisions
)

type ExtractionStatus string

const (
	StatusPending    ExtractionStatus = "pending"
	StatusProcessing ExtractionStatus = "processing"
	StatusDone       ExtractionStatus = "done"
	StatusFailed     ExtractionStatus = "failed"
)

const MinJDLength = 50

// JobDescription is the parsed representation of a target JD.
// Invariant: RawText is immutable after creation — LLM extracts from it, never replaces it.
// Invariant: CandidateID enforces multi-tenancy — JDs belong to a candidate.
type JobDescription struct {
	ID               uuid.UUID
	CandidateID      uuid.UUID
	CompanyID        *uuid.UUID      // nullable — matched from company registry or nil
	RawText          string          // Invariant: never modified after creation
	NormalisedText   string          // whitespace-cleaned, URL-stripped
	SenioritySignal  SenioritySignal
	ArchExpectation  ArchExpectation
	ExtractionStatus ExtractionStatus
	ExtractionError  string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func New(candidateID uuid.UUID, rawText string) (*JobDescription, error) {
	if candidateID == uuid.Nil {
		return nil, errors.New("jd: CandidateID is required")
	}
	if len(rawText) < MinJDLength {
		return nil, errors.New("jd: RawText too short to be a valid job description")
	}
	now := time.Now().UTC()
	return &JobDescription{
		ID:               uuid.New(),
		CandidateID:      candidateID,
		RawText:          rawText,
		SenioritySignal:  SeniorityUnknown,
		ArchExpectation:  ArchNone,
		ExtractionStatus: StatusPending,
		CreatedAt:        now,
		UpdatedAt:        now,
	}, nil
}

func (j *JobDescription) Validate() error {
	if j.CandidateID == uuid.Nil {
		return errors.New("jd: CandidateID is required")
	}
	if len(j.RawText) < MinJDLength {
		return errors.New("jd: RawText too short")
	}
	return nil
}

func (j *JobDescription) MarkDone(normalisedText string, seniority SenioritySignal, arch ArchExpectation, companyID *uuid.UUID) {
	j.NormalisedText = normalisedText
	j.SenioritySignal = seniority
	j.ArchExpectation = arch
	j.CompanyID = companyID
	j.ExtractionStatus = StatusDone
	j.ExtractionError = ""
	j.UpdatedAt = time.Now().UTC()
}

func (j *JobDescription) MarkFailed(errMsg string) {
	j.ExtractionStatus = StatusFailed
	j.ExtractionError = errMsg
	j.UpdatedAt = time.Now().UTC()
}

// JDSkill is a skill required or preferred in the JD.
// Invariant: Weight across all required skills for a JD sums to 1.0.
// Invariant: MinRequiredScore in [1, 10].
type JDSkill struct {
	ID               uuid.UUID
	JDID             uuid.UUID
	SkillID          uuid.UUID
	SkillName        string  // denormalised for display
	IsRequired       bool
	MinRequiredScore int     // [1, 10]
	Weight           float64 // fraction of required skills — sums to 1.0 across required
	CreatedAt        time.Time
}

func (j *JDSkill) Validate() error {
	if j.JDID == uuid.Nil || j.SkillID == uuid.Nil {
		return errors.New("jd_skill: JDID and SkillID are required")
	}
	if j.MinRequiredScore < 1 || j.MinRequiredScore > 10 {
		return errors.New("jd_skill: MinRequiredScore must be in [1, 10]")
	}
	if j.Weight < 0 || j.Weight > 1 {
		return errors.New("jd_skill: Weight must be in [0, 1]")
	}
	return nil
}
