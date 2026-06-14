-- Revert: make assessment_id NOT NULL again
ALTER TABLE coach_sessions
    ALTER COLUMN assessment_id SET NOT NULL;
