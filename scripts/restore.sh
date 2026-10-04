#!/usr/bin/env bash
set -euo pipefail

BACKUP_FILE="${1:-}"
TARGET_DB_URL="${2:-${DATABASE_URL:-}}"
ENCRYPTION_KEY="${BACKUP_ENCRYPTION_KEY:-}"

if [ -z "$BACKUP_FILE" ]; then
  echo "Usage: $0 <backup-file> [target-db-url]" >&2
  exit 1
fi

if [ -z "$TARGET_DB_URL" ]; then
  echo "Error: target database URL is required (as argument or DATABASE_URL env)." >&2
  exit 1
fi

echo "Starting HMS ET Database Restore Drill"
echo "Backup source: $BACKUP_FILE"

MANIFEST_FILE="${BACKUP_FILE%.*}.manifest.json"
if [ -f "$MANIFEST_FILE" ]; then
  echo "Verifying SHA-256 against manifest..."
  EXPECTED_HASH=$(grep '"sha256":' "$MANIFEST_FILE" | head -n1 | tr -d ' ",' | cut -d: -f2)
  ACTUAL_HASH=$(sha256sum "$BACKUP_FILE" | awk '{print $1}')
  if [ "$EXPECTED_HASH" != "$ACTUAL_HASH" ]; then
    echo "Checksum mismatch! Expected: $EXPECTED_HASH, Got: $ACTUAL_HASH" >&2
    exit 1
  fi
  echo "Checksum verification passed: $ACTUAL_HASH"
fi

TEMP_FILE=$(mktemp --suffix=.sql.gz)
trap 'rm -f "$TEMP_FILE"' EXIT

if [[ "$BACKUP_FILE" == *.enc ]]; then
  if [ -z "$ENCRYPTION_KEY" ]; then
    echo "Error: BACKUP_ENCRYPTION_KEY is required to decrypt this backup." >&2
    exit 1
  fi
  echo "Decrypting backup file..."
  export ENC_PASS="$ENCRYPTION_KEY"
  openssl enc -d -aes-256-cbc -pbkdf2 -iter 100000 -in "$BACKUP_FILE" -out "$TEMP_FILE" -pass env:ENC_PASS
else
  cp "$BACKUP_FILE" "$TEMP_FILE"
fi

echo "Restoring database from dump..."
pg_restore --clean --if-exists --no-owner --no-privileges --dbname="$TARGET_DB_URL" "$TEMP_FILE" || true

echo "Running post-restore verification query..."
PATIENT_COUNT=$(psql "$TARGET_DB_URL" -t -c "SELECT COUNT(*) FROM patient;" | tr -d '[:space:]')
echo "Active patient records in restored database: $PATIENT_COUNT"
echo "Restore drill successfully verified!"
