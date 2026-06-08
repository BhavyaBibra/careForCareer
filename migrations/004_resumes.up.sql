CREATE TYPE resume_source AS ENUM ('pdf', 'github', 'none');
CREATE TYPE extraction_status AS ENUM ('pending', 'processing', 'done', 'failed');

CREATE TABLE resumes (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    candidate_id        UUID NOT NULL REFERENCES candidates(id) ON DELETE CASCADE,
    source_type         resume_source NOT NULL,
    storage_key         TEXT,
    github_url          TEXT,
    raw_text            TEXT,
    extraction_status   extraction_status NOT NULL DEFAULT 'pending',
    extraction_error    TEXT,
    version             INT NOT NULL DEFAULT 1,
    parse_attempts      INT NOT NULL DEFAULT 0 CHECK (parse_attempts <= 3),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT resume_source_check CHECK (
        (source_type = 'pdf'    AND storage_key IS NOT NULL) OR
        (source_type = 'github' AND github_url  IS NOT NULL) OR
        (source_type = 'none')
    )
);

CREATE INDEX idx_resumes_candidate_id ON resumes(candidate_id);
CREATE INDEX idx_resumes_status ON resumes(extraction_status)
    WHERE extraction_status != 'done';

ALTER TABLE resumes ENABLE ROW LEVEL SECURITY;
CREATE POLICY resumes_own ON resumes
    USING (candidate_id IN (
        SELECT id FROM candidates
        WHERE user_id = current_setting('app.current_user_id', true)::UUID
    ));
