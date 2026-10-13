-- Migration 067: Clinical Encounters Parity, Performance Indexes & Atomic Discharge Bed Repair

-- 1. Optimized indexes for encounter filtering, status lookups, and care team delegations
CREATE INDEX IF NOT EXISTS idx_encounter_kind_status ON encounter(kind, status, admitted_at DESC);
CREATE INDEX IF NOT EXISTS idx_encounter_patient_kind ON encounter(patient_id, kind, status);
CREATE INDEX IF NOT EXISTS idx_encounter_doctor_kind ON encounter(doctor_id, kind, status);
CREATE INDEX IF NOT EXISTS idx_encounter_care_team_staff ON encounter_care_team(staff_id, encounter_id) WHERE revoked_at IS NULL;

-- 2. Performance indexes for child registers
CREATE INDEX IF NOT EXISTS idx_encounter_diagnosis_category ON encounter_diagnosis(encounter_id, category, status);
CREATE INDEX IF NOT EXISTS idx_encounter_procedure_performed ON encounter_procedure(encounter_id, performed_at DESC);
CREATE INDEX IF NOT EXISTS idx_opd_follow_up_status ON opd_follow_up(encounter_id, status);

-- 3. Historical data repair: Reconcile orphaned bed assignments for concluded stays
-- If an encounter was marked discharged, close any orphaned open bed assignments.
-- Preserves maintenance and unavailable states: lack of occupancy alone does not establish readiness.
UPDATE bed_assignment
SET assigned_to = COALESCE(e.discharged_at, clock_timestamp())
FROM encounter e
WHERE bed_assignment.encounter_id = e.id
  AND bed_assignment.assigned_to IS NULL
  AND e.status = 'discharged';

-- 4. Enable ON UPDATE CASCADE for doctor reassignment across case and encounters
ALTER TABLE encounter DROP CONSTRAINT IF EXISTS encounter_case_id_patient_id_doctor_id_fkey;
ALTER TABLE encounter ADD CONSTRAINT encounter_case_id_patient_id_doctor_id_fkey
  FOREIGN KEY (case_id, patient_id, doctor_id)
  REFERENCES patient_case(id, patient_id, doctor_id)
  ON UPDATE CASCADE;
