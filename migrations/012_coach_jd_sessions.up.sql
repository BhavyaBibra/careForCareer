-- Allow JD-aware coach sessions that don't require a prior readiness assessment.
-- assessment_id becomes nullable; FK preserved for rows that do have one.
ALTER TABLE coach_sessions
    ALTER COLUMN assessment_id DROP NOT NULL,
    ALTER COLUMN assessment_id DROP DEFAULT;

-- Replace FK constraint to allow NULL while still enforcing referential integrity for non-null rows
ALTER TABLE coach_sessions DROP CONSTRAINT IF EXISTS coach_sessions_assessment_id_fkey;
ALTER TABLE coach_sessions
    ADD CONSTRAINT coach_sessions_assessment_id_fkey
    FOREIGN KEY (assessment_id) REFERENCES readiness_assessments(id) ON DELETE CASCADE
    NOT VALID;
