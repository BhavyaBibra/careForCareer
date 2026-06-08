CREATE TYPE skill_category AS ENUM (
    'dsa', 'backend', 'system_design', 'architecture',
    'domain', 'language', 'devops'
);

CREATE TABLE skills (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name       TEXT NOT NULL UNIQUE,
    category   skill_category NOT NULL,
    aliases    TEXT[] NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_skills_name ON skills(name);
CREATE INDEX idx_skills_aliases ON skills USING GIN(aliases);

CREATE TABLE candidate_skills (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    candidate_id    UUID NOT NULL REFERENCES candidates(id) ON DELETE CASCADE,
    skill_id        UUID NOT NULL REFERENCES skills(id) ON DELETE RESTRICT,
    resume_id       UUID NOT NULL REFERENCES resumes(id) ON DELETE CASCADE,
    score           SMALLINT NOT NULL CHECK (score BETWEEN 1 AND 10),
    confidence      NUMERIC(3,2) NOT NULL CHECK (confidence BETWEEN 0.00 AND 1.00),
    evidence_source TEXT NOT NULL,
    raw_evidence    TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(candidate_id, skill_id, resume_id)
);

CREATE INDEX idx_candidate_skills_candidate ON candidate_skills(candidate_id);
CREATE INDEX idx_candidate_skills_skill ON candidate_skills(skill_id);

ALTER TABLE candidate_skills ENABLE ROW LEVEL SECURITY;
CREATE POLICY candidate_skills_own ON candidate_skills
    USING (candidate_id IN (
        SELECT id FROM candidates
        WHERE user_id = current_setting('app.current_user_id', true)::UUID
    ));
