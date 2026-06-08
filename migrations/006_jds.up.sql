CREATE TYPE seniority_signal AS ENUM ('junior', 'mid', 'senior', 'staff', 'unknown');
CREATE TYPE arch_expectation AS ENUM ('none', 'team', 'org', 'platform');

CREATE TABLE job_descriptions (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    candidate_id        UUID NOT NULL REFERENCES candidates(id) ON DELETE CASCADE,
    company_id          UUID,
    raw_text            TEXT NOT NULL,
    normalised_text     TEXT,
    seniority_signal    seniority_signal NOT NULL DEFAULT 'unknown',
    arch_expectation    arch_expectation NOT NULL DEFAULT 'none',
    extraction_status   extraction_status NOT NULL DEFAULT 'pending',
    extraction_error    TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_jds_candidate_id ON job_descriptions(candidate_id);
CREATE INDEX idx_jds_status ON job_descriptions(extraction_status)
    WHERE extraction_status != 'done';

CREATE TABLE jd_skills (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    jd_id               UUID NOT NULL REFERENCES job_descriptions(id) ON DELETE CASCADE,
    skill_id            UUID NOT NULL REFERENCES skills(id) ON DELETE RESTRICT,
    skill_name          TEXT NOT NULL DEFAULT '',
    is_required         BOOLEAN NOT NULL DEFAULT TRUE,
    min_required_score  SMALLINT NOT NULL CHECK (min_required_score BETWEEN 1 AND 10),
    weight              NUMERIC(5,4) NOT NULL DEFAULT 0,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(jd_id, skill_id)
);

CREATE INDEX idx_jd_skills_jd ON jd_skills(jd_id);
