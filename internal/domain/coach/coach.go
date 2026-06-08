package coach

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"careergps/internal/domain/candidate"
	"careergps/internal/domain/gap"
	"careergps/internal/domain/readiness"
	"careergps/internal/domain/roadmap"
)

type MessageRole string

const (
	RoleUser      MessageRole = "user"
	RoleAssistant MessageRole = "assistant"
)

// PrepPlanSummary is the coach's view of the current roadmap state.
type PrepPlanSummary struct {
	InterviewDate time.Time
	TotalDays     int
	DaysRemaining int
	CurrentDay    int
	TodaysTasks   []string
	IsStale       bool
	OverdueDays   int
}

// CoachContext is assembled fresh on every coach turn from deterministic data sources.
// Never derived from LLM memory. This is the grounding contract — the coach cannot
// invent scores or gaps; it always refers back to this context.
type CoachContext struct {
	Candidate   *candidate.Candidate
	GapAnalysis *gap.GapAnalysis
	Assessment  *readiness.ReadinessAssessment
	PrepPlan    *PrepPlanSummary
}

// BuildPrepSummary derives the PrepPlanSummary from a Roadmap.
func BuildPrepSummary(r *roadmap.Roadmap) *PrepPlanSummary {
	if r == nil {
		return nil
	}
	tasks := r.TodaysTasks()
	taskTitles := make([]string, 0, len(tasks))
	for _, t := range tasks {
		taskTitles = append(taskTitles, t.Title)
	}
	return &PrepPlanSummary{
		InterviewDate: r.InterviewDate,
		TotalDays:     r.TotalDays(),
		DaysRemaining: r.DaysRemaining(),
		CurrentDay:    r.CurrentDay(),
		TodaysTasks:   taskTitles,
		IsStale:       r.IsStale(),
		OverdueDays:   r.OverdueDays(),
	}
}

// AssembleSystemPrompt builds the grounded context block injected before every LLM call.
// This is rebuilt fresh on every coach turn from deterministic sources.
func AssembleSystemPrompt(ctx *CoachContext) string {
	var b strings.Builder

	b.WriteString("You are CareerGPS Coach — a precise, direct career advisor for software engineers in India.\n")
	b.WriteString("You have access to this candidate's actual skill data, gap analysis, and prep plan.\n")
	b.WriteString("RULES: Never invent skill scores. Never hallucinate company interview patterns. Always refer to the data below.\n\n")

	b.WriteString("[CANDIDATE CONTEXT]\n")
	if ctx.Candidate != nil {
		fmt.Fprintf(&b, "Tier: %s (inferred from %d YOE)\n",
			ctx.Candidate.InferredTier.TierLabel(), ctx.Candidate.YearsExperience)
		if ctx.Candidate.CurrentCompany != "" {
			fmt.Fprintf(&b, "Current Company: %s\n", ctx.Candidate.CurrentCompany)
		}
	}
	if ctx.Assessment != nil {
		fmt.Fprintf(&b, "Readiness Score: %.0f%% (engine: %s)\n",
			ctx.Assessment.CompositeScore, ctx.Assessment.EngineVersion)
	}

	if ctx.GapAnalysis != nil {
		b.WriteString("\n[GAP SUMMARY]\n")
		critical := ctx.GapAnalysis.CriticalGaps(3)
		onTrack := ctx.GapAnalysis.OnTrackSkills()
		for _, g := range critical {
			fmt.Fprintf(&b, "CRITICAL: %s (need %d, have %d, gap %d)\n",
				g.SkillName, g.JDRequiredScore, g.CandidateScore, g.Gap)
		}
		for _, g := range onTrack {
			surplus := g.JDRequiredScore - g.CandidateScore
			fmt.Fprintf(&b, "ON TRACK: %s (surplus %d)\n", g.SkillName, -surplus)
		}
		fmt.Fprintf(&b, "Overall fit: %s (confidence: %.0f%%)\n",
			ctx.GapAnalysis.FitLevel, ctx.GapAnalysis.Confidence*100)
	}

	if ctx.PrepPlan != nil {
		b.WriteString("\n[PREP PLAN]\n")
		fmt.Fprintf(&b, "Day %d of %d — Interview: %s (%d days remaining)\n",
			ctx.PrepPlan.CurrentDay, ctx.PrepPlan.TotalDays,
			ctx.PrepPlan.InterviewDate.Format("Jan 2, 2006"),
			ctx.PrepPlan.DaysRemaining)
		if ctx.PrepPlan.IsStale {
			fmt.Fprintf(&b, "WARNING: Plan is %d days stale. Candidate is behind schedule.\n",
				ctx.PrepPlan.OverdueDays)
		}
		if len(ctx.PrepPlan.TodaysTasks) > 0 {
			fmt.Fprintf(&b, "Today's tasks: %s\n", strings.Join(ctx.PrepPlan.TodaysTasks, "; "))
		}
	}

	b.WriteString("\nAnswer questions grounded in the data above. Be concise and actionable.\n")
	return b.String()
}

// CoachSession binds a candidate to a coach conversation.
// ContextSnapshot preserves the CoachContext at session start for audit.
// V1: message history is per-session; not persisted across logins.
type CoachSession struct {
	ID              uuid.UUID
	CandidateID     uuid.UUID
	AssessmentID    uuid.UUID
	ContextSnapshot string    // JSON of CoachContext at session start
	CreatedAt       time.Time
	ExpiresAt       time.Time // 24h inactivity expiry
}

func NewSession(candidateID, assessmentID uuid.UUID, contextSnapshot string) (*CoachSession, error) {
	if candidateID == uuid.Nil || assessmentID == uuid.Nil {
		return nil, errors.New("coach_session: CandidateID and AssessmentID are required")
	}
	now := time.Now().UTC()
	return &CoachSession{
		ID:              uuid.New(),
		CandidateID:     candidateID,
		AssessmentID:    assessmentID,
		ContextSnapshot: contextSnapshot,
		CreatedAt:       now,
		ExpiresAt:       now.Add(24 * time.Hour),
	}, nil
}

func (s *CoachSession) IsExpired() bool {
	return time.Now().UTC().After(s.ExpiresAt)
}

// CoachMessage is a single turn in the conversation.
// TokenCost and LatencyMs are tracked for cost analysis.
// Invariant: Content is never empty for a completed message.
type CoachMessage struct {
	ID        uuid.UUID
	SessionID uuid.UUID
	Role      MessageRole
	Content   string
	TokenCost int
	LatencyMs int64
	CreatedAt time.Time
}

func NewMessage(sessionID uuid.UUID, role MessageRole, content string) (*CoachMessage, error) {
	if sessionID == uuid.Nil {
		return nil, errors.New("coach_message: SessionID is required")
	}
	if content == "" {
		return nil, errors.New("coach_message: Content is required")
	}
	return &CoachMessage{
		ID:        uuid.New(),
		SessionID: sessionID,
		Role:      role,
		Content:   content,
		CreatedAt: time.Now().UTC(),
	}, nil
}
