CREATE TABLE roadmaps (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    candidate_id    UUID NOT NULL REFERENCES candidates(id) ON DELETE CASCADE,
    assessment_id   UUID NOT NULL REFERENCES readiness_assessments(id) ON DELETE CASCADE,
    interview_date  DATE NOT NULL,
    plan_start_date DATE NOT NULL DEFAULT CURRENT_DATE,
    daily_tasks     JSONB NOT NULL DEFAULT '[]',
    narrative       TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_roadmaps_candidate ON roadmaps(candidate_id);
CREATE INDEX idx_roadmaps_assessment ON roadmaps(assessment_id);
