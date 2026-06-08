package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

// Task type constants — used as Asynq task type strings.
const (
	TaskParseResumePDF    = "resume:parse_pdf"
	TaskParseGitHub       = "resume:parse_github"
	TaskExtractJD         = "jd:extract"
	TaskRunGapAnalysis    = "assessment:gap_analysis"
	TaskRunReadiness      = "assessment:readiness"
	TaskBuildRoadmap      = "roadmap:build"
	TaskGenerateNarrative = "roadmap:narrative"
	TaskGenerateQuestions = "company:questions"
	TaskExplainScore      = "assessment:explain"
)

// Queue names — mapped to Asynq priority weights.
const (
	QueueCritical = "critical" // priority 10
	QueueHigh     = "high"     // priority 7
	QueueDefault  = "default"  // priority 5
	QueueLow      = "low"      // priority 2
)

// ParseResumePDFPayload is the job payload for PDF text extraction.
type ParseResumePDFPayload struct {
	ResumeID    string `json:"resume_id"`
	CandidateID string `json:"candidate_id"`
	StorageKey  string `json:"storage_key"`
	Version     int    `json:"version"`
}

// ParseGitHubPayload is the job payload for GitHub signal extraction.
type ParseGitHubPayload struct {
	ResumeID    string `json:"resume_id"`
	CandidateID string `json:"candidate_id"`
	GitHubURL   string `json:"github_url"`
}

// ExtractJDPayload is the job payload for JD parsing.
type ExtractJDPayload struct {
	JDID        string `json:"jd_id"`
	CandidateID string `json:"candidate_id"`
}

// GapAnalysisPayload is the job payload for gap computation.
type GapAnalysisPayload struct {
	CandidateID string `json:"candidate_id"`
	ResumeID    string `json:"resume_id"`
	JDID        string `json:"jd_id"`
}

// IdempotencyKey builds a deduplicated task ID.
// Format: {job_type}:{entity_id}:{version}
func IdempotencyKey(jobType, entityID string, version int) string {
	return fmt.Sprintf("%s:%s:v%d", jobType, entityID, version)
}

// NewTask creates an Asynq task with a JSON payload.
func NewTask(taskType string, payload interface{}) (*asynq.Task, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("worker: marshal payload for %s: %w", taskType, err)
	}
	return asynq.NewTask(taskType, b), nil
}

// HandleParseResumePDF is a placeholder handler. Real implementation reads PDF from S3,
// extracts text, calls LLM for skill extraction, persists CandidateSkills.
func HandleParseResumePDF(log *zap.Logger) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		var p ParseResumePDFPayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return fmt.Errorf("parse_pdf: unmarshal: %w", err)
		}
		log.Info("processing resume pdf", zap.String("resume_id", p.ResumeID))
		// Full implementation in infrastructure/parser
		return nil
	}
}

// HandleExtractJD is a placeholder handler. Real implementation calls LLM for JD skill extraction.
func HandleExtractJD(log *zap.Logger) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		var p ExtractJDPayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return fmt.Errorf("extract_jd: unmarshal: %w", err)
		}
		log.Info("extracting jd", zap.String("jd_id", p.JDID))
		return nil
	}
}

// HandleRunGapAnalysis runs gap analysis after resume + JD are both parsed.
func HandleRunGapAnalysis(log *zap.Logger) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		var p GapAnalysisPayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return fmt.Errorf("gap_analysis: unmarshal: %w", err)
		}
		log.Info("running gap analysis",
			zap.String("candidate_id", p.CandidateID),
			zap.String("jd_id", p.JDID),
		)
		return nil
	}
}
