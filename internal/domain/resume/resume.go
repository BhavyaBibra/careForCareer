package resume

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type SourceType string

const (
	SourcePDF    SourceType = "pdf"
	SourceGitHub SourceType = "github"
	SourceNone   SourceType = "none"
)

type ExtractionStatus string

const (
	StatusPending    ExtractionStatus = "pending"
	StatusProcessing ExtractionStatus = "processing"
	StatusDone       ExtractionStatus = "done"
	StatusFailed     ExtractionStatus = "failed"
)

const MaxParseAttempts = 3

// Resume holds the candidate's raw and parsed experience signal.
// Invariant: for SourcePDF, StorageKey must be set.
// Invariant: for SourceGitHub, GitHubURL must be set.
// Invariant: ParseAttempts <= MaxParseAttempts — enforced before job enqueue.
// Invariant: RawText is immutable once set — re-parsing creates a new version.
type Resume struct {
	ID               uuid.UUID
	CandidateID      uuid.UUID
	SourceType       SourceType
	StorageKey       string           // S3 object key — empty for github/none
	GitHubURL        string           // empty for pdf/none
	RawText          string           // populated after Stage 1 extraction
	ExtractionStatus ExtractionStatus
	ExtractionError  string           // last error if status=failed
	Version          int              // monotonically increasing per candidate
	ParseAttempts    int              // Invariant: <= MaxParseAttempts
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func New(candidateID uuid.UUID, sourceType SourceType, storageKey, githubURL string, version int) (*Resume, error) {
	r := &Resume{
		ID:               uuid.New(),
		CandidateID:      candidateID,
		SourceType:       sourceType,
		StorageKey:       storageKey,
		GitHubURL:        githubURL,
		ExtractionStatus: StatusPending,
		Version:          version,
		ParseAttempts:    0,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}
	if err := r.Validate(); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *Resume) Validate() error {
	if r.CandidateID == uuid.Nil {
		return errors.New("resume: CandidateID is required")
	}
	switch r.SourceType {
	case SourcePDF:
		if r.StorageKey == "" {
			return errors.New("resume: StorageKey required for pdf source")
		}
	case SourceGitHub:
		if r.GitHubURL == "" {
			return errors.New("resume: GitHubURL required for github source")
		}
	case SourceNone:
		// valid — fresh grad zero-resume path
	default:
		return errors.New("resume: invalid SourceType")
	}
	if r.ParseAttempts > MaxParseAttempts {
		return errors.New("resume: ParseAttempts exceeds maximum")
	}
	return nil
}

// MarkProcessing transitions state to processing and increments attempt counter.
func (r *Resume) MarkProcessing() error {
	if r.ParseAttempts >= MaxParseAttempts {
		return errors.New("resume: max parse attempts reached")
	}
	r.ExtractionStatus = StatusProcessing
	r.ParseAttempts++
	r.UpdatedAt = time.Now().UTC()
	return nil
}

// MarkDone finalises extraction with the parsed raw text.
func (r *Resume) MarkDone(rawText string) {
	r.RawText = rawText
	r.ExtractionStatus = StatusDone
	r.ExtractionError = ""
	r.UpdatedAt = time.Now().UTC()
}

// MarkFailed records the error and reverts status.
func (r *Resume) MarkFailed(errMsg string) {
	r.ExtractionStatus = StatusFailed
	r.ExtractionError = errMsg
	r.UpdatedAt = time.Now().UTC()
}

// CanRetry returns true if another parse attempt is allowed.
func (r *Resume) CanRetry() bool {
	return r.ParseAttempts < MaxParseAttempts
}
