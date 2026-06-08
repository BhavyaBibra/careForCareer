CREATE TABLE llm_conversations (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    candidate_id    UUID REFERENCES candidates(id) ON DELETE SET NULL,
    provider        TEXT NOT NULL,
    model           TEXT NOT NULL,
    prompt_hash     TEXT NOT NULL,
    job_type        TEXT,
    input_tokens    INT NOT NULL DEFAULT 0,
    output_tokens   INT NOT NULL DEFAULT 0,
    latency_ms      BIGINT NOT NULL DEFAULT 0,
    cache_hit       BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_llm_conv_candidate ON llm_conversations(candidate_id);
CREATE INDEX idx_llm_conv_prompt_hash ON llm_conversations(prompt_hash);
CREATE INDEX idx_llm_conv_created ON llm_conversations(created_at);

CREATE TYPE job_status AS ENUM ('pending', 'processing', 'done', 'failed', 'dead');

CREATE TABLE async_jobs (
    id               UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    idempotency_key  TEXT NOT NULL UNIQUE,
    queue            TEXT NOT NULL,
    job_type         TEXT NOT NULL,
    payload          JSONB NOT NULL DEFAULT '{}',
    status           job_status NOT NULL DEFAULT 'pending',
    attempts         INT NOT NULL DEFAULT 0,
    max_attempts     INT NOT NULL DEFAULT 3,
    last_error       TEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    scheduled_at     TIMESTAMPTZ,
    completed_at     TIMESTAMPTZ
);

CREATE INDEX idx_async_jobs_status ON async_jobs(status) WHERE status != 'done';
CREATE INDEX idx_async_jobs_queue ON async_jobs(queue, status);
CREATE INDEX idx_async_jobs_idempotency ON async_jobs(idempotency_key);
