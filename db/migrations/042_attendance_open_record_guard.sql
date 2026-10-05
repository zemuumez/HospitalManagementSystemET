-- Refuse migration if legacy overlapping open records require administrator
-- reconciliation; never silently discard or close retained attendance.
CREATE UNIQUE INDEX attendance_one_open_per_staff
 ON attendance_record(staff_id) WHERE check_out_at IS NULL;
ALTER TABLE attendance_record ADD CONSTRAINT attendance_review_requires_checkout
 CHECK (approval_status NOT IN ('submitted','approved') OR check_out_at IS NOT NULL);
