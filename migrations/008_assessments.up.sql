CREATE TABLE gap_analyses (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    candidate_id    UUID NOT NULL REFERENCES candidates(id) ON DELETE CASCADE,
    jd_id           UUID NOT NULL REFERENCES job_descriptions(id) ON DELETE CASCADE,
    gaps_json       JSONB NOT NULL,
    aggregate_gap   NUMERIC(5,2) NOT NULL,
    confidence      NUMERIC(3,2) NOT NULL,
    fit_level       TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_gap_analyses_candidate ON gap_analyses(candidate_id);
CREATE INDEX idx_gap_analyses_jd ON gap_analyses(jd_id);

CREATE TABLE readiness_assessments (
    id               UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    candidate_id     UUID NOT NULL REFERENCES candidates(id) ON DELETE CASCADE,
    jd_id            UUID NOT NULL REFERENCES job_descriptions(id) ON DELETE CASCADE,
    gap_analysis_id  UUID NOT NULL REFERENCES gap_analyses(id) ON DELETE RESTRICT,
    tier             SMALLINT NOT NULL,
    engine_version   TEXT NOT NULL,
    composite_score  NUMERIC(5,2) NOT NULL CHECK (composite_score BETWEEN 0 AND 100),
    components_json  JSONB NOT NULL,
    weights_json     JSONB NOT NULL,
    input_snapshot   JSONB NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_assessments_candidate ON readiness_assessments(candidate_id);
CREATE INDEX idx_assessments_jd ON readiness_assessments(jd_id);
CREATE INDEX idx_assessments_engine ON readiness_assessments(engine_version);
