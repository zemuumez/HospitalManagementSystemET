-- Migration 052: Doctor schedules, holidays, and lunch breaks parity
-- Provides dedicated relational storage matching Laravel HMS models:
-- Schedule, ScheduleDay, DoctorHoliday, and LunchBreak.

-- 1. Doctor holidays (full-day leave / absences with reason)
CREATE TABLE IF NOT EXISTS doctor_holiday (
  id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  doctor_id    text NOT NULL REFERENCES doctor_profile(user_id) ON DELETE CASCADE,
  holiday_date date NOT NULL,
  name         text NOT NULL DEFAULT '' CHECK (length(name) <= 200),
  created_by   text NOT NULL REFERENCES "user"(id),
  created_at   timestamptz NOT NULL DEFAULT now(),
  UNIQUE(doctor_id, holiday_date)
);
CREATE INDEX IF NOT EXISTS idx_doctor_holiday_doctor_date ON doctor_holiday(doctor_id, holiday_date);

-- 2. Doctor lunch breaks (recurring daily or single-date interval)
CREATE TABLE IF NOT EXISTS doctor_lunch_break (
  id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  doctor_id   text NOT NULL REFERENCES doctor_profile(user_id) ON DELETE CASCADE,
  break_from  time NOT NULL,
  break_to    time NOT NULL,
  every_day   boolean NOT NULL DEFAULT true,
  break_date  date,
  created_by  text NOT NULL REFERENCES "user"(id),
  created_at  timestamptz NOT NULL DEFAULT now(),
  CHECK (break_to > break_from),
  CHECK ((every_day = true AND break_date IS NULL) OR (every_day = false AND break_date IS NOT NULL))
);
CREATE INDEX IF NOT EXISTS idx_doctor_break_doctor ON doctor_lunch_break(doctor_id);
