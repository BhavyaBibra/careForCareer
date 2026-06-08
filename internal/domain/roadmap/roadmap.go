package roadmap

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// DailyTask represents one study item in the prep plan.
type DailyTask struct {
	Title           string  `json:"title"`
	Description     string  `json:"description"`
	SkillID         *uuid.UUID `json:"skill_id,omitempty"`
	EstimatedHours  float64 `json:"estimated_hours"`
	Priority        int     `json:"priority"` // 1 = highest
	Category        string  `json:"category"` // "dsa" | "system_design" | "backend" | etc.
}

// DayPlan maps a day number to its tasks.
type DayPlan struct {
	Day   int         `json:"day"`
	Tasks []DailyTask `json:"tasks"`
}

// Roadmap is the time-aware, interview-date-driven prep plan.
// Invariant: InterviewDate must be in the future at creation time.
// Invariant: DayPlans are ordered by Day ascending.
// Narrative is nullable — populated asynchronously by LLM enrichment (Stage 7).
type Roadmap struct {
	ID            uuid.UUID
	CandidateID   uuid.UUID
	AssessmentID  uuid.UUID
	InterviewDate time.Time
	PlanStartDate time.Time
	DayPlans      []DayPlan
	Narrative     *string // LLM-generated summary; nil until Stage 7 completes
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func New(candidateID, assessmentID uuid.UUID, interviewDate time.Time, dayPlans []DayPlan) (*Roadmap, error) {
	now := time.Now().UTC()
	today := now.Truncate(24 * time.Hour)
	if !interviewDate.After(today) {
		return nil, errors.New("roadmap: InterviewDate must be in the future")
	}
	return &Roadmap{
		ID:            uuid.New(),
		CandidateID:   candidateID,
		AssessmentID:  assessmentID,
		InterviewDate: interviewDate,
		PlanStartDate: today,
		DayPlans:      dayPlans,
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

func (r *Roadmap) Validate() error {
	if r.CandidateID == uuid.Nil || r.AssessmentID == uuid.Nil {
		return errors.New("roadmap: CandidateID and AssessmentID are required")
	}
	if r.InterviewDate.IsZero() {
		return errors.New("roadmap: InterviewDate is required")
	}
	return nil
}

// TotalDays returns the number of days in the prep plan.
func (r *Roadmap) TotalDays() int {
	return len(r.DayPlans)
}

// DaysRemaining returns days until the interview from today.
func (r *Roadmap) DaysRemaining() int {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	d := int(r.InterviewDate.Sub(today).Hours() / 24)
	if d < 0 {
		return 0
	}
	return d
}

// CurrentDay returns which plan day the candidate is on (1-indexed).
func (r *Roadmap) CurrentDay() int {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	d := int(today.Sub(r.PlanStartDate).Hours()/24) + 1
	if d < 1 {
		return 1
	}
	if d > r.TotalDays() {
		return r.TotalDays()
	}
	return d
}

// IsStale returns true when the candidate is behind schedule.
func (r *Roadmap) IsStale() bool {
	return r.CurrentDay() > r.DaysRemaining()+1
}

// OverdueDays returns how many days behind schedule the candidate is.
func (r *Roadmap) OverdueDays() int {
	overdue := r.CurrentDay() - r.DaysRemaining() - 1
	if overdue < 0 {
		return 0
	}
	return overdue
}

// TodaysTasks returns the tasks for the current plan day.
func (r *Roadmap) TodaysTasks() []DailyTask {
	currentDay := r.CurrentDay()
	for _, dp := range r.DayPlans {
		if dp.Day == currentDay {
			return dp.Tasks
		}
	}
	return nil
}

// SetNarrative sets the LLM-generated narrative after async enrichment.
func (r *Roadmap) SetNarrative(narrative string) {
	r.Narrative = &narrative
	r.UpdatedAt = time.Now().UTC()
}
