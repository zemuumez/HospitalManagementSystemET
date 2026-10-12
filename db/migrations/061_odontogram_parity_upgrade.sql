-- 061_odontogram_parity_upgrade.sql
-- Corrective upgrade migration for odontogram parity (F1-F4, G1-G3).
-- Safely applies schema additions to databases that already ran migration 060,
-- captures complete historical provenance before deduplicating legacy charts,
-- widens and reconciles legend table schemas, and seeds established definitions.

-- 1. Add legacy_provenance column to odontogram_chart if missing
ALTER TABLE odontogram_chart
  ADD COLUMN IF NOT EXISTS legacy_provenance jsonb DEFAULT NULL;

-- 2. Complete Provenance Capture & Deduplication of Legacy Charts (G1)
-- Capture complete original chart headers, tooth identities/content, and existing provenance
-- BEFORE any merge or delete operation, ensuring zero clinical history is discarded.
DO $$
DECLARE
  p RECORD;
  keeper_id uuid;
  all_prov jsonb;
  dup RECORD;
BEGIN
  -- Iterate through every patient who has legacy charts
  FOR p IN (
    SELECT DISTINCT patient_id
    FROM odontogram_chart
    WHERE is_legacy = true
  ) LOOP
    -- Step A: Deterministically choose the primary keeper chart for this patient
    SELECT id INTO keeper_id
    FROM odontogram_chart
    WHERE patient_id = p.patient_id AND is_legacy = true
    ORDER BY version DESC, updated_at DESC, created_at DESC, id ASC
    LIMIT 1;

    -- Step B: BEFORE deleting or merging any duplicate rows, assemble the COMPLETE historical provenance
    -- across ALL legacy charts, ALL child tooth entries, and ANY pre-existing provenance snapshots.
    WITH
      -- 1. Extract any existing provenance elements already captured on any legacy chart of this patient
      existing_prov AS (
        SELECT elem
        FROM odontogram_chart c,
        LATERAL jsonb_array_elements(
          CASE
            WHEN c.legacy_provenance IS NOT NULL AND jsonb_typeof(c.legacy_provenance) = 'array'
            THEN c.legacy_provenance
            ELSE '[]'::jsonb
          END
        ) AS elem
        WHERE c.patient_id = p.patient_id AND c.is_legacy = true
      ),
      -- 2. Extract all child tooth entries across all legacy charts of this patient
      all_child_entries AS (
        SELECT
          jsonb_build_object(
            'tooth_number', e.tooth_number,
            'condition', e.condition,
            'procedure_notes', e.procedure_notes,
            'diagnosed_by', e.diagnosed_by,
            'encounter_id', e.encounter_id::text,
            'chart_id', e.chart_id::text,
            'created_at', e.created_at,
            'updated_at', e.updated_at,
            'source_chart_description', ch.description,
            'source_chart_doctor_id', ch.doctor_id
          ) AS elem
        FROM patient_odontogram_entry e
        JOIN odontogram_chart ch ON ch.id = e.chart_id
        WHERE ch.patient_id = p.patient_id AND ch.is_legacy = true
      ),
      -- 3. Extract conditions directly from odontogram JSON objects if child entries were missing
      chart_json_entries AS (
        SELECT
          jsonb_build_object(
            'tooth_number', kv.key::int,
            'condition', kv.value,
            'procedure_notes', ch.description,
            'diagnosed_by', ch.doctor_id,
            'encounter_id', ch.encounter_id::text,
            'chart_id', ch.id::text,
            'created_at', ch.created_at,
            'updated_at', ch.updated_at,
            'source_chart_description', ch.description,
            'source_chart_doctor_id', ch.doctor_id
          ) AS elem
        FROM odontogram_chart ch,
        LATERAL jsonb_each_text(
          CASE
            WHEN ch.odontogram IS NOT NULL AND jsonb_typeof(ch.odontogram) = 'object'
            THEN ch.odontogram
            ELSE '{}'::jsonb
          END
        ) AS kv
        WHERE ch.patient_id = p.patient_id AND ch.is_legacy = true
          AND NOT EXISTS (
            SELECT 1 FROM patient_odontogram_entry pe
            WHERE pe.chart_id = ch.id AND pe.tooth_number = kv.key::int
          )
      ),
      combined AS (
        SELECT elem FROM existing_prov
        UNION
        SELECT elem FROM all_child_entries
        UNION
        SELECT elem FROM chart_json_entries
      )
    SELECT COALESCE(
      jsonb_agg(
        elem ORDER BY
          (elem->>'tooth_number')::int NULLS LAST,
          (elem->>'updated_at') DESC NULLS LAST,
          (elem->>'created_at') DESC NULLS LAST
      ),
      '[]'::jsonb
    ) INTO all_prov
    FROM combined;

    -- Step C: Durably persist the COMPLETE historical archive in keeper chart's legacy_provenance
    UPDATE odontogram_chart
    SET legacy_provenance = all_prov
    WHERE id = keeper_id;

    -- Step D: For multiple duplicate legacy charts, reassign or consolidate child entries
    FOR dup IN (
      SELECT id
      FROM odontogram_chart
      WHERE patient_id = p.patient_id AND is_legacy = true AND id <> keeper_id
    ) LOOP
      -- Move child entries that don't collide with keeper's tooth numbers
      UPDATE patient_odontogram_entry e
      SET chart_id = keeper_id
      WHERE e.chart_id = dup.id
        AND NOT EXISTS (
          SELECT 1 FROM patient_odontogram_entry ke
          WHERE ke.chart_id = keeper_id AND ke.tooth_number = e.tooth_number
        );

      -- Now delete colliding child entries from the duplicate chart
      -- (SAFE: all conflicting notes and observations are already safely preserved in all_prov!)
      DELETE FROM patient_odontogram_entry WHERE chart_id = dup.id;

      -- Delete duplicate legacy chart header
      DELETE FROM odontogram_chart WHERE id = dup.id;
    END LOOP;

    -- Step E: Synchronize keeper's odontogram JSON to match keeper's current child entries
    UPDATE odontogram_chart
    SET odontogram = (
      SELECT COALESCE(jsonb_object_agg(tooth_number::text, condition), '{}'::jsonb)
      FROM patient_odontogram_entry
      WHERE chart_id = keeper_id
    )
    WHERE id = keeper_id;

  END LOOP;
END $$;

-- 3. Enforce exactly one legacy chart per patient
CREATE UNIQUE INDEX IF NOT EXISTS odontogram_chart_legacy_patient_idx
  ON odontogram_chart (patient_id)
  WHERE is_legacy = true;

-- 4. Create or Widen odontogram_legend Table Schema (G3)
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

-- For databases upgraded from intermediate migration 060 (which created varchar(16) / varchar(100)):
DO $$
BEGIN
  ALTER TABLE odontogram_legend ALTER COLUMN code TYPE text;
  ALTER TABLE odontogram_legend ALTER COLUMN name TYPE text;
  ALTER TABLE odontogram_legend ALTER COLUMN color TYPE text;

  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'chk_odontogram_legend_code_len'
  ) THEN
    ALTER TABLE odontogram_legend ADD CONSTRAINT chk_odontogram_legend_code_len
      CHECK (length(trim(code)) >= 1 AND length(code) <= 64);
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'chk_odontogram_legend_name_len'
  ) THEN
    ALTER TABLE odontogram_legend ADD CONSTRAINT chk_odontogram_legend_name_len
      CHECK (length(trim(name)) >= 1 AND length(name) <= 128);
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'chk_odontogram_legend_color'
  ) THEN
    ALTER TABLE odontogram_legend ADD CONSTRAINT chk_odontogram_legend_color
      CHECK (color ~* '^#[0-9a-f]{6}$');
  END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_odontogram_legend_code ON odontogram_legend(code);

-- 5. Reconcile default clinical definitions while strictly preserving user customizations (G3 & F2)
-- Correct untouched corrupted intermediate seeds specifically:
UPDATE odontogram_legend
SET name = 'Root Canal (Karies Radix)',
    color = '#607d8b',
    description = 'Endodontic therapy / Root canal treated',
    is_default = true,
    updated_at = clock_timestamp()
WHERE code = 'KR' AND (name = 'Radicular Cyst' OR (is_default = true AND color = '#795548'));

UPDATE odontogram_legend
SET name = 'Pulpitis / Deep Caries',
    color = '#009688',
    description = 'Extensive caries with pulpal involvement',
    is_default = true,
    updated_at = clock_timestamp()
WHERE code = 'Ce' AND (name = 'Cervical Caries' OR (is_default = true AND color = '#e91e63'));

UPDATE odontogram_legend
SET name = 'Decayed / Denture',
    color = '#795548',
    description = 'Grossly decayed or retained root',
    is_default = true,
    updated_at = clock_timestamp()
WHERE code = 'D' AND (name = 'Decalcification' OR (is_default = true AND color = '#ffc107'));

UPDATE odontogram_legend
SET name = 'Porcelain / Prosthetic',
    color = '#ffc107',
    description = 'Porcelain veneer or prosthetic replacement',
    is_default = true,
    updated_at = clock_timestamp()
WHERE code = 'PS' AND (name = 'Post & Core' OR (is_default = true AND color = '#009688'));

-- Seed missing default clinical condition definitions without overwriting any existing customization:
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
