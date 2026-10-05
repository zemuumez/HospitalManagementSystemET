-- Migration 040: Secure Private Attachments and Storage Management
-- Enforces size and MIME type limits, non-predictable cryptographic access tokens,
-- and strict ownership/audit tracking for medical attachments.

CREATE TABLE secure_attachment (
  id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  token           varchar(64) NOT NULL UNIQUE CHECK (length(token) >= 32),
  file_name       varchar(255) NOT NULL CHECK (length(trim(file_name)) >= 1),
  mime_type       varchar(100) NOT NULL CHECK (
    mime_type IN (
      'application/pdf',
      'image/jpeg',
      'image/png',
      'image/webp',
      'text/plain',
      'text/csv',
      'application/dicom',
      'application/zip'
    )
  ),
  file_size_bytes bigint NOT NULL CHECK (file_size_bytes > 0 AND file_size_bytes <= 26214400), -- 25MB max
  storage_path    text NOT NULL UNIQUE CHECK (length(trim(storage_path)) >= 1),
  sha256_hash     varchar(64) NOT NULL CHECK (length(sha256_hash) = 64),
  uploader_id     text NOT NULL REFERENCES "user"(id),
  patient_id      uuid REFERENCES patient(id) ON DELETE SET NULL,
  encounter_id    uuid REFERENCES encounter(id) ON DELETE SET NULL,
  is_public       boolean NOT NULL DEFAULT false,
  created_at      timestamptz NOT NULL DEFAULT clock_timestamp()
);

CREATE INDEX idx_secure_attachment_token ON secure_attachment(token);
CREATE INDEX idx_secure_attachment_patient ON secure_attachment(patient_id);
CREATE INDEX idx_secure_attachment_encounter ON secure_attachment(encounter_id);
CREATE INDEX idx_secure_attachment_uploader ON secure_attachment(uploader_id);
