CREATE TYPE company_tier AS ENUM (
    'faang', 'global_product', 'unicorn', 'mid_startup', 'service'
);

CREATE TABLE companies (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name            TEXT NOT NULL UNIQUE,
    tier            company_tier NOT NULL,
    india_bar_notes TEXT NOT NULL DEFAULT '',
    website         TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_companies_tier ON companies(tier);
CREATE INDEX idx_companies_name ON companies(name);

-- Interview patterns: admin CRUD, no code deploy needed to update.
CREATE TABLE company_patterns (
    id                          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id                  UUID NOT NULL UNIQUE REFERENCES companies(id) ON DELETE CASCADE,
    interview_rounds            JSONB NOT NULL DEFAULT '[]',
    focus_areas                 TEXT[] NOT NULL DEFAULT '{}',
    typical_rejection_reasons   TEXT[] NOT NULL DEFAULT '{}',
    dsa_difficulty              TEXT NOT NULL DEFAULT 'lc_hard',
    notes                       TEXT,
    updated_by                  TEXT NOT NULL DEFAULT 'system',
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- FK back-references now that companies table exists
ALTER TABLE job_descriptions
    ADD CONSTRAINT fk_jd_company
    FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE SET NULL;

ALTER TABLE candidate_target_companies
    ADD CONSTRAINT fk_target_company
    FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE CASCADE;
