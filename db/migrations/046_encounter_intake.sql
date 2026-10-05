-- Preserve original registration fields atomically with admission.
ALTER TABLE encounter ADD COLUMN intake jsonb NOT NULL DEFAULT '{}'::jsonb CHECK (jsonb_typeof(intake) = 'object');
