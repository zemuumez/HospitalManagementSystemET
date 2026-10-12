-- 061_odontogram_parity_upgrade.sql
-- Corrective upgrade migration for odontogram parity (F1 & F2).
-- Safely applies schema additions to databases that already ran migration 060,
-- cleans up legacy chart duplicates, captures available legacy provenance,
-- and seeds the established clinical legend definitions.

-- 1. Add legacy_provenance column to odontogram_chart if missing
ALTER TABLE odontogram_chart
  ADD COLUMN IF NOT EXISTS legacy_provenance jsonb DEFAULT NULL;

-- 2. Deduplicate legacy charts per patient (handles artifacts from earlier 9eaa114 compatibility writes)
DO $$
DECLARE
  p RECORD;
  primary_id uuid;
  dup RECORD;
BEGIN
  FOR p IN (
    SELECT patient_id, COUNT(*) AS cnt
    FROM odontogram_chart
    WHERE is_legacy = true
    GROUP BY patient_id
    HAVING COUNT(*) > 1
  ) LOOP
    -- Choose the latest chart as primary keeper
    SELECT id INTO primary_id
    FROM odontogram_chart
    WHERE patient_id = p.patient_id AND is_legacy = true
    ORDER BY updated_at DESC, created_at DESC, version DESC
    LIMIT 1;

    -- For all other duplicate legacy charts for this patient:
    FOR dup IN (
      SELECT id, odontogram
      FROM odontogram_chart
      WHERE patient_id = p.patient_id AND is_legacy = true AND id <> primary_id
    ) LOOP
      -- Move child entries from duplicate to primary if tooth does not already exist in primary
      UPDATE patient_odontogram_entry e
      SET chart_id = primary_id
      WHERE e.chart_id = dup.id
        AND NOT EXISTS (
          SELECT 1 FROM patient_odontogram_entry pe
          WHERE pe.chart_id = primary_id AND pe.tooth_number = e.tooth_number
        );

      -- Delete remaining conflicting child entries in the duplicate chart
      DELETE FROM patient_odontogram_entry WHERE chart_id = dup.id;

      -- Update primary chart's odontogram JSON to reflect all child entries
      UPDATE odontogram_chart
      SET odontogram = (
        SELECT COALESCE(jsonb_object_agg(tooth_number::text, condition), '{}'::jsonb)
        FROM patient_odontogram_entry
        WHERE chart_id = primary_id
      )
      WHERE id = primary_id;

      -- Delete duplicate legacy chart
      DELETE FROM odontogram_chart WHERE id = dup.id;
    END LOOP;
  END LOOP;
END $$;

-- 3. Enforce exactly one legacy chart per patient
CREATE UNIQUE INDEX IF NOT EXISTS odontogram_chart_legacy_patient_idx
  ON odontogram_chart (patient_id)
  WHERE is_legacy = true;

-- 4. Backfill legacy_provenance snapshot for existing legacy charts if not yet captured
UPDATE odontogram_chart c
SET legacy_provenance = (
  SELECT COALESCE(
    jsonb_agg(
      jsonb_build_object(
        'tooth_number', e.tooth_number,
        'condition', e.condition,
        'procedure_notes', e.procedure_notes,
        'diagnosed_by', e.diagnosed_by,
        'encounter_id', e.encounter_id::text,
        'created_at', e.created_at,
        'updated_at', e.updated_at
      ) ORDER BY e.tooth_number
    ),
    '[]'::jsonb
  )
  FROM patient_odontogram_entry e
  WHERE e.chart_id = c.id
)
WHERE c.is_legacy = true AND c.legacy_provenance IS NULL;

-- 5. Create odontogram_legend table for durable clinical definitions
CREATE TABLE IF NOT EXISTS odontogram_legend (
  id          text PRIMARY KEY,
  code        text NOT NULL UNIQUE CHECK (length(trim(code)) >= 1 AND length(code) <= 64),
  name        text NOT NULL CHECK (length(trim(name)) >= 1 AND length(name) <= 128),
  color       text NOT NULL CHECK (color ~* '^#[0-9a-f]{6}$'),
  description text NOT NULL DEFAULT '',
  is_default  boolean NOT NULL DEFAULT false,
  created_at  timestamptz NOT NULL DEFAULT clock_timestamp(),
  updated_at  timestamptz NOT NULL DEFAULT clock_timestamp()
);

CREATE INDEX IF NOT EXISTS idx_odontogram_legend_code ON odontogram_legend(code);

-- 6. Seed and reconcile the 11 established default clinical condition definitions (F2)
INSERT INTO odontogram_legend (id, code, name, color, description, is_default)
VALUES
  ('leg-healthy', 'Healthy', 'Healthy Tooth',              '#ffffff', 'Sound tooth surface with no pathology', true),
  ('leg-k',       'K',       'Caries / Cavity',           '#e91e63', 'Active dental caries lesion', true),
  ('leg-c',       'C',       'Crown',                     '#3f51b5', 'Artificial full crown or cap', true),
  ('leg-ce',      'Ce',      'Pulpitis / Deep Caries',    '#009688', 'Extensive caries with pulpal involvement', true),
  ('leg-d',       'D',       'Decayed / Denture',         '#795548', 'Grossly decayed or retained root', true),
  ('leg-kr',      'KR',      'Root Canal (Karies Radix)', '#607d8b', 'Endodontic therapy / Root canal treated', true),
  ('leg-ps',      'PS',      'Porcelain / Prosthetic',    '#ffc107', 'Porcelain veneer or prosthetic replacement', true),
  ('leg-ip',      'IP',      'Implant',                   '#673ab7', 'Titanium or ceramic dental implant', true),
  ('leg-x',       'X',       'Missing / Extracted',       '#2196f3', 'Congenitally missing or extracted tooth', true),
  ('leg-f',       'F',       'Composite Filling',         '#10b981', 'Tooth-colored composite restoration', true),
  ('leg-b',       'B',       'Bridge',                    '#f97316', 'Fixed partial denture / pontic unit', true)
ON CONFLICT (code) DO NOTHING;
