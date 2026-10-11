-- 060_odontogram_chart_parity.sql
-- Create odontogram_chart table, link patient_odontogram_entry to chart,
-- migrate pre-existing tooth entries into legacy charts, and replace per-patient
-- tooth uniqueness with chart-scoped uniqueness.

-- 1. Create odontogram_chart table
CREATE TABLE IF NOT EXISTS odontogram_chart (
  id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  patient_id   uuid NOT NULL REFERENCES patient(id) ON DELETE CASCADE,
  doctor_id    text NOT NULL REFERENCES "user"(id),
  encounter_id uuid REFERENCES encounter(id) ON DELETE SET NULL,
  description  text NOT NULL CHECK (length(trim(description)) >= 1 AND length(description) <= 2000),
  odontogram   jsonb NOT NULL DEFAULT '{}'::jsonb,
  version      integer NOT NULL DEFAULT 1 CHECK (version >= 1),
  created_by   text NOT NULL REFERENCES "user"(id),
  updated_by   text NOT NULL REFERENCES "user"(id),
  created_at        timestamptz NOT NULL DEFAULT clock_timestamp(),
  updated_at        timestamptz NOT NULL DEFAULT clock_timestamp(),
  is_legacy         boolean NOT NULL DEFAULT false,
  legacy_provenance jsonb DEFAULT NULL
);

CREATE INDEX IF NOT EXISTS idx_odontogram_chart_patient ON odontogram_chart(patient_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_odontogram_chart_doctor ON odontogram_chart(doctor_id);
CREATE INDEX IF NOT EXISTS idx_odontogram_chart_created ON odontogram_chart(created_at DESC);
CREATE UNIQUE INDEX IF NOT EXISTS odontogram_chart_legacy_patient_idx ON odontogram_chart (patient_id) WHERE is_legacy = true;

-- 2. Add chart_id to patient_odontogram_entry (nullable temporarily for backfill)
ALTER TABLE patient_odontogram_entry
  ADD COLUMN IF NOT EXISTS chart_id uuid REFERENCES odontogram_chart(id) ON DELETE CASCADE;

-- 3. Backfill pre-existing rows in patient_odontogram_entry where chart_id IS NULL
DO $$
DECLARE
  r RECORD;
  new_chart_id uuid;
  doc_id text;
  chart_desc text;
  chart_enc uuid;
  min_created timestamptz;
  max_updated timestamptz;
  teeth_json jsonb;
  prov_json jsonb;
BEGIN
  FOR r IN (
    SELECT DISTINCT patient_id
    FROM patient_odontogram_entry
    WHERE chart_id IS NULL
  ) LOOP
    -- Gather aggregated values and complete per-tooth provenance for this patient
    SELECT
      (ARRAY_AGG(diagnosed_by ORDER BY created_at DESC))[1],
      COALESCE(NULLIF((ARRAY_AGG(procedure_notes) FILTER (WHERE length(trim(procedure_notes)) > 0))[1], ''), 'Legacy examination record'),
      (ARRAY_AGG(encounter_id) FILTER (WHERE encounter_id IS NOT NULL))[1],
      MIN(created_at),
      MAX(updated_at),
      COALESCE(jsonb_object_agg(tooth_number::text, condition), '{}'::jsonb),
      COALESCE(jsonb_agg(jsonb_build_object(
        'tooth_number', tooth_number,
        'condition', condition,
        'procedure_notes', procedure_notes,
        'diagnosed_by', diagnosed_by,
        'encounter_id', encounter_id::text,
        'created_at', created_at,
        'updated_at', updated_at
      ) ORDER BY tooth_number), '[]'::jsonb)
    INTO
      doc_id, chart_desc, chart_enc, min_created, max_updated, teeth_json, prov_json
    FROM patient_odontogram_entry
    WHERE patient_id = r.patient_id AND chart_id IS NULL;

    -- Ensure a valid doctor_id exists (must exist in "user")
    IF doc_id IS NULL OR NOT EXISTS (SELECT 1 FROM "user" WHERE id = doc_id) THEN
      SELECT id INTO doc_id FROM "user" WHERE email LIKE '%admin%' LIMIT 1;
      IF doc_id IS NULL THEN
        SELECT id INTO doc_id FROM "user" LIMIT 1;
      END IF;
    END IF;

    -- Create identified legacy chart with immutable provenance snapshot
    INSERT INTO odontogram_chart (
      id, patient_id, doctor_id, encounter_id, description,
      odontogram, version, created_by, updated_by, created_at, updated_at, is_legacy, legacy_provenance
    ) VALUES (
      gen_random_uuid(), r.patient_id, doc_id, chart_enc, chart_desc,
      teeth_json, 1, doc_id, doc_id, min_created, max_updated, true, prov_json
    ) RETURNING id INTO new_chart_id;

    -- Link the patient's existing tooth rows to this legacy chart
    UPDATE patient_odontogram_entry
    SET chart_id = new_chart_id
    WHERE patient_id = r.patient_id AND chart_id IS NULL;
  END LOOP;
END $$;

-- 4. Now that existing entries are linked, enforce NOT NULL on chart_id
ALTER TABLE patient_odontogram_entry
  ALTER COLUMN chart_id SET NOT NULL;

-- 5. Replace per-patient tooth uniqueness with chart-scoped uniqueness
ALTER TABLE patient_odontogram_entry
  DROP CONSTRAINT IF EXISTS patient_odontogram_entry_patient_id_tooth_number_key;

ALTER TABLE patient_odontogram_entry
  DROP CONSTRAINT IF EXISTS patient_odontogram_entry_chart_tooth_key;

ALTER TABLE patient_odontogram_entry
  ADD CONSTRAINT patient_odontogram_entry_chart_tooth_key UNIQUE (chart_id, tooth_number);

CREATE INDEX IF NOT EXISTS idx_odontogram_entry_chart ON patient_odontogram_entry(chart_id);

-- 6. Relax condition check to preserve all 11 condition labels losslessly and support custom legends
ALTER TABLE patient_odontogram_entry
  DROP CONSTRAINT IF EXISTS patient_odontogram_entry_condition_check;

ALTER TABLE patient_odontogram_entry
  ADD CONSTRAINT patient_odontogram_entry_condition_check
  CHECK (length(condition) >= 1 AND length(condition) <= 64);

-- 7. Durable Odontogram Legends (Persistent Clinical Definitions)
CREATE TABLE IF NOT EXISTS odontogram_legend (
  id          varchar(64) PRIMARY KEY,
  code        varchar(16) NOT NULL UNIQUE,
  name        varchar(100) NOT NULL,
  color       varchar(32) NOT NULL,
  description text NOT NULL DEFAULT '',
  is_default  boolean NOT NULL DEFAULT false,
  created_at  timestamptz NOT NULL DEFAULT clock_timestamp(),
  updated_at  timestamptz NOT NULL DEFAULT clock_timestamp()
);

CREATE INDEX IF NOT EXISTS idx_odontogram_legend_code ON odontogram_legend(code);

-- Seed 11 default clinical condition legends
INSERT INTO odontogram_legend (id, code, name, color, description, is_default)
VALUES
  ('leg-healthy', 'Healthy', 'Healthy / Sound Tooth', '#ffffff', 'Intact, non-carious anatomical tooth structure without restorations', true),
  ('leg-k',       'K',       'Caries (Cavity)',          '#ef4444', 'Active carious lesion or enamel cavitation requiring treatment', true),
  ('leg-c',       'C',       'Crown',                    '#3b82f6', 'Full coverage artificial crown restoration (metal / ceramic)', true),
  ('leg-ce',      'Ce',      'Cervical Caries',          '#f59e0b', 'Cervical / root surface carious lesion at gingival margin', true),
  ('leg-d',       'D',       'Decalcification',          '#eab308', 'Enamel demineralization / white spot lesion without frank cavity', true),
  ('leg-kr',      'KR',      'Radicular Cyst',           '#8b5cf6', 'Periapical / radicular odontogenic cystic lesion', true),
  ('leg-ps',      'PS',      'Post & Core',              '#6366f1', 'Endodontic post and foundation core restoration', true),
  ('leg-ip',      'IP',      'Implant',                  '#06b6d4', 'Endosseous dental fixture implant fixture integrated in alveolar bone', true),
  ('leg-x',       'X',       'Extracted / Missing',      '#94a3b8', 'Extracted, congenitally missing, or unerupted tooth space', true),
  ('leg-f',       'F',       'Composite Filling',        '#10b981', 'Tooth-colored composite restoration placed and functional', true),
  ('leg-b',       'B',       'Bridge',                   '#f97316', 'Fixed partial denture abutment or pontic replacement unit', true)
ON CONFLICT (code) DO NOTHING;

