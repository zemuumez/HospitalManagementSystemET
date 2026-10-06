-- Preserve existing shift identities and make the default independent of its name.
ALTER TABLE attendance_shift
  ADD COLUMN code text NOT NULL DEFAULT ('SHIFT-' || upper(substr(replace(gen_random_uuid()::text,'-',''),1,26))),
  ADD COLUMN is_default boolean NOT NULL DEFAULT false;
UPDATE attendance_shift SET code='DS', is_default=active WHERE id='11111111-1111-1111-1111-111111111111';
UPDATE attendance_shift SET code='NS' WHERE id='22222222-2222-2222-2222-222222222222';
ALTER TABLE attendance_shift ADD CONSTRAINT attendance_shift_code_valid
  CHECK (code ~ '^[A-Z0-9][A-Z0-9_-]{0,31}$');
ALTER TABLE attendance_shift ADD CONSTRAINT attendance_shift_default_active CHECK (NOT is_default OR active);
CREATE UNIQUE INDEX attendance_shift_code_unique ON attendance_shift(code);
CREATE UNIQUE INDEX attendance_shift_single_default ON attendance_shift(is_default) WHERE is_default;
