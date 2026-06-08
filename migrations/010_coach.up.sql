CREATE TABLE coach_sessions (
    id                UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    candidate_id      UUID NOT NULL REFERENCES candidates(id) ON DELETE CASCADE,
    assessment_id     UUID NOT NULL REFERENCES readiness_assessments(id) ON DELETE CASCADE,
    context_snapshot  JSONB NOT NULL DEFAULT '{}',
    expires_at        TIMESTAMPTZ NOT NULL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_coach_sessions_candidate ON coach_sessions(candidate_id);
CREATE INDEX idx_coach_sessions_expires ON coach_sessions(expires_at);

CREATE TABLE coach_messages (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id   UUID NOT NULL REFERENCES coach_sessions(id) ON DELETE CASCADE,
    role         TEXT NOT NULL CHECK (role IN ('user', 'assistant')),
    content      TEXT NOT NULL,
    token_cost   INT NOT NULL DEFAULT 0,
    latency_ms   BIGINT NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_coach_messages_session ON coach_messages(session_id);
CREATE INDEX idx_coach_messages_session_time ON coach_messages(session_id, created_at);
