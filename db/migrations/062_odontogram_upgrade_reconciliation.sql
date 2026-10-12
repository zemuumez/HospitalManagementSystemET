-- 062_odontogram_upgrade_reconciliation.sql
-- Corrective upgrade migration for odontogram parity (H1, H2, H3).
-- Safely applies schema corrections to databases where migration 061 has already been applied,
-- reconciles intermediate-060 column widths and check constraints,
-- performs non-destructive complete legacy chart header and child observation provenance archiving,
-- and reconciles default clinical legend seeds using field-aware preservation of user customizations.
--
-- NOTE ON HISTORICAL DATA RECOVERY (H1):
-- In environments where migration 061 (from commit 796bfc4) was already executed prior to this migration,
-- duplicate chart headers or colliding child teeth that were deleted by old 061 cannot be reconstructed
-- from surviving database tables alone; recovering those previously deleted records requires restoring from
-- a pre-upgrade database backup or external audit log. Migration 062 archives all surviving headers, child entries,
-- and provenance records losslessly, prevents any further data loss, and ensures full schema and seed integrity.

-- 1. Ensure legacy_provenance column exists on odontogram_chart
ALTER TABLE odontogram_chart
  ADD COLUMN IF NOT EXISTS legacy_provenance jsonb DEFAULT NULL;

-- 2. Schema Widening & Contract Reconciliation (H1 & G3)
-- Reconcile narrow column definitions from intermediate migration 060 (varchar(16) / varchar(100)):
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

-- 3. Comprehensive Provenance Archiving & Duplicate Header/Tooth Consolidation (H2)
-- Archives complete source headers (including identity, timestamps, attribution, encounter, description,
-- version, and JSON) and child row identities (id) before any consolidation.
DO $$
DECLARE
  p RECORD;
  keeper_id uuid;
  all_prov jsonb;
  dup RECORD;
  existing_prov jsonb;
  all_headers jsonb;
  all_children jsonb;
BEGIN
  -- For every patient who has legacy charts:
  FOR p IN (
    SELECT DISTINCT patient_id
    FROM odontogram_chart
    WHERE is_legacy = true
  ) LOOP
    -- Step A: Deterministically identify primary keeper chart
    SELECT id INTO keeper_id
    FROM odontogram_chart
    WHERE patient_id = p.patient_id AND is_legacy = true
    ORDER BY version DESC, updated_at DESC, created_at DESC, id ASC
    LIMIT 1;

    -- Step B: Assemble complete provenance across all legacy charts for this patient
    -- 1. All source chart headers (preserves empty duplicates, header descriptions, timestamps, attribution)
    SELECT COALESCE(
      jsonb_agg(
        jsonb_build_object(
          'type', 'chart_header',
          'chart_id', ch.id::text,
          'patient_id', ch.patient_id::text,
          'doctor_id', ch.doctor_id,
          'encounter_id', ch.encounter_id::text,
          'description', ch.description,
          'odontogram', ch.odontogram,
          'version', ch.version,
          'created_by', ch.created_by,
          'updated_by', ch.updated_by,
          'created_at', ch.created_at,
          'updated_at', ch.updated_at
        ) ORDER BY ch.created_at ASC, ch.id ASC
      ),
      '[]'::jsonb
    ) INTO all_headers
    FROM odontogram_chart ch
    WHERE ch.patient_id = p.patient_id AND ch.is_legacy = true;

    -- 2. All child tooth entries across all legacy charts (including child primary key 'id')
    SELECT COALESCE(
      jsonb_agg(
        jsonb_build_object(
          'type', 'tooth_entry',
          'id', e.id::text,
          'chart_id', e.chart_id::text,
          'tooth_number', e.tooth_number,
          'condition', e.condition,
          'procedure_notes', e.procedure_notes,
          'diagnosed_by', e.diagnosed_by,
          'encounter_id', e.encounter_id::text,
          'created_at', e.created_at,
          'updated_at', e.updated_at
        ) ORDER BY e.tooth_number ASC, e.created_at ASC, e.id ASC
      ),
      '[]'::jsonb
    ) INTO all_children
    FROM patient_odontogram_entry e
    JOIN odontogram_chart ch ON ch.id = e.chart_id
    WHERE ch.patient_id = p.patient_id AND ch.is_legacy = true;

    -- 3. Pre-existing provenance entries already recorded on any legacy chart of this patient
    SELECT COALESCE(
      jsonb_agg(elem),
      '[]'::jsonb
    ) INTO existing_prov
    FROM odontogram_chart ch,
    LATERAL jsonb_array_elements(
      CASE
        WHEN ch.legacy_provenance IS NOT NULL AND jsonb_typeof(ch.legacy_provenance) = 'array'
        THEN ch.legacy_provenance
        ELSE '[]'::jsonb
      END
    ) AS elem
    WHERE ch.patient_id = p.patient_id AND ch.is_legacy = true;

    -- Combine headers, children, and prior provenance losslessly, deduplicating identical JSON objects
    SELECT COALESCE(jsonb_agg(DISTINCT elem), '[]'::jsonb)
    INTO all_prov
    FROM jsonb_array_elements(all_headers || all_children || existing_prov) AS elem;

    -- Step C: If duplicate legacy charts exist, consolidate child teeth and remove duplicate headers
    FOR dup IN (
      SELECT id FROM odontogram_chart
      WHERE patient_id = p.patient_id AND is_legacy = true AND id <> keeper_id
    ) LOOP
      -- Move non-conflicting teeth to keeper
      UPDATE patient_odontogram_entry e
      SET chart_id = keeper_id
      WHERE e.chart_id = dup.id
        AND NOT EXISTS (
          SELECT 1 FROM patient_odontogram_entry ke
          WHERE ke.chart_id = keeper_id AND ke.tooth_number = e.tooth_number
        );

      -- Delete colliding child teeth from duplicate (already archived in all_prov!)
      DELETE FROM patient_odontogram_entry WHERE chart_id = dup.id;

      -- Delete duplicate chart header (already archived in all_prov!)
      DELETE FROM odontogram_chart WHERE id = dup.id;
    END LOOP;

    -- Step D: Store the comprehensive provenance on keeper chart
    UPDATE odontogram_chart
    SET legacy_provenance = all_prov,
        odontogram = (
          SELECT COALESCE(jsonb_object_agg(tooth_number::text, condition), '{}'::jsonb)
          FROM patient_odontogram_entry
          WHERE chart_id = keeper_id
        )
    WHERE id = keeper_id;

  END LOOP;
END $$;

-- Enforce exactly one legacy chart per patient
CREATE UNIQUE INDEX IF NOT EXISTS odontogram_chart_legacy_patient_idx
  ON odontogram_chart (patient_id)
  WHERE is_legacy = true;

-- 4. Field-Aware Seed Reconciliation Preserving User Customizations (H3)
-- Reconciles default clinical definitions for KR, Ce, D, PS while strictly preserving
-- customized names, colors, and descriptions on each field individually.
-- Custom rows (is_default = false) are never modified.

-- KR: Root Canal (Karies Radix)
UPDATE odontogram_legend
SET
  name = CASE
    WHEN name IN ('Radicular Cyst') THEN 'Root Canal (Karies Radix)'
    ELSE name
  END,
  color = CASE
    WHEN lower(color) IN ('#8b5cf6', '#795548') THEN '#607d8b'
    ELSE color
  END,
  description = CASE
    WHEN description IN ('Periapical / radicular odontogenic cystic lesion', 'Periapical / radicular cyst', '')
      THEN 'Endodontic therapy / Root canal treated'
    ELSE description
  END,
  updated_at = clock_timestamp()
WHERE code = 'KR' AND is_default = true;

-- Ce: Pulpitis / Deep Caries
UPDATE odontogram_legend
SET
  name = CASE
    WHEN name IN ('Cervical Caries') THEN 'Pulpitis / Deep Caries'
    ELSE name
  END,
  color = CASE
    WHEN lower(color) IN ('#f59e0b', '#e91e63') THEN '#009688'
    ELSE color
  END,
  description = CASE
    WHEN description IN ('Cervical / root surface carious lesion at gingival margin', 'Cervical / root caries', '')
      THEN 'Extensive caries with pulpal involvement'
    ELSE description
  END,
  updated_at = clock_timestamp()
WHERE code = 'Ce' AND is_default = true;

-- D: Decayed / Denture
UPDATE odontogram_legend
SET
  name = CASE
    WHEN name IN ('Decalcification') THEN 'Decayed / Denture'
    ELSE name
  END,
  color = CASE
    WHEN lower(color) IN ('#eab308', '#ffc107') THEN '#795548'
    ELSE color
  END,
  description = CASE
    WHEN description IN ('Enamel demineralization / white spot lesion without frank cavity', '')
      THEN 'Grossly decayed or retained root'
    ELSE description
  END,
  updated_at = clock_timestamp()
WHERE code = 'D' AND is_default = true;

-- PS: Porcelain / Prosthetic
UPDATE odontogram_legend
SET
  name = CASE
    WHEN name IN ('Post & Core') THEN 'Porcelain / Prosthetic'
    ELSE name
  END,
  color = CASE
    WHEN lower(color) IN ('#6366f1', '#009688') THEN '#ffc107'
    ELSE color
  END,
  description = CASE
    WHEN description IN ('Endodontic post and foundation core restoration', 'Endodontic post and core', '')
      THEN 'Porcelain veneer or prosthetic replacement'
    ELSE description
  END,
  updated_at = clock_timestamp()
WHERE code = 'PS' AND is_default = true;

-- Seed missing default clinical condition definitions without overwriting existing codes:
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
