package company

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Tier string

const (
	TierFAANG         Tier = "faang"
	TierGlobalProduct Tier = "global_product"
	TierUnicorn       Tier = "unicorn"
	TierMidStartup    Tier = "mid_startup"
	TierService       Tier = "service"
)

// Company is a seed entity — created via admin API or migrations.
// IndiaBarNotes captures India-specific calibration.
type Company struct {
	ID            uuid.UUID
	Name          string
	Tier          Tier
	IndiaBarNotes string // e.g. "Google BLR L3 bar: LC Hard 3/5, system design required"
	Website       string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (c *Company) Validate() error {
	if c.Name == "" {
		return errors.New("company: Name is required")
	}
	if c.Tier == "" {
		return errors.New("company: Tier is required")
	}
	return nil
}

// RoundDescriptor describes one interview round.
type RoundDescriptor struct {
	Order       int    `json:"order"`
	Name        string `json:"name"`         // e.g. "System Design", "Bar Raiser"
	DurationMin int    `json:"duration_min"`
	Focus       string `json:"focus"`
}

// CompanyPattern captures interview structure for a company.
// Admin-updatable via PUT /api/v1/admin/companies/{id}/patterns — no code deploy needed.
// Invariant: InterviewRounds, FocusAreas, TypicalRejectionReasons stored as JSONB.
type CompanyPattern struct {
	ID                      uuid.UUID
	CompanyID               uuid.UUID
	InterviewRounds         []RoundDescriptor
	FocusAreas              []string // e.g. ["kafka", "distributed-systems", "lld"]
	TypicalRejectionReasons []string // e.g. ["weak system design", "poor communication"]
	DSADifficulty           string   // "lc_easy_medium" | "lc_hard" | "custom"
	Notes                   string
	UpdatedBy               string
	CreatedAt               time.Time
	UpdatedAt               time.Time
}

func (cp *CompanyPattern) Validate() error {
	if cp.CompanyID == uuid.Nil {
		return errors.New("company_pattern: CompanyID is required")
	}
	if cp.DSADifficulty == "" {
		return errors.New("company_pattern: DSADifficulty is required")
	}
	return nil
}

// InterviewRoundsJSON serialises rounds to JSON for DB storage.
func (cp *CompanyPattern) InterviewRoundsJSON() ([]byte, error) {
	return json.Marshal(cp.InterviewRounds)
}
