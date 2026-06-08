CREATE TABLE candidates (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id             UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    years_experience    INT NOT NULL CHECK (years_experience >= 0),
    inferred_tier       SMALLINT NOT NULL CHECK (inferred_tier BETWEEN 0 AND 4),
    tier_explanation    TEXT NOT NULL DEFAULT '',
    current_company     TEXT,
    current_comp_inr    BIGINT,
    target_comp_inr     BIGINT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_candidates_user_id ON candidates(user_id);

ALTER TABLE candidates ENABLE ROW LEVEL SECURITY;
CREATE POLICY candidates_own_row ON candidates
    USING (user_id = current_setting('app.current_user_id', true)::UUID);

CREATE TABLE candidate_target_companies (
    candidate_id UUID NOT NULL REFERENCES candidates(id) ON DELETE CASCADE,
    company_id   UUID NOT NULL,
    PRIMARY KEY  (candidate_id, company_id)
);
