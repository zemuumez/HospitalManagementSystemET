#!/usr/bin/env bash
set -euo pipefail

BACKUP_DIR="${BACKUP_DIR:-backups}"
RETENTION_DAYS="${RETENTION_DAYS:-30}"
ENCRYPTION_KEY="${BACKUP_ENCRYPTION_KEY:-}"

mkdir -p "$BACKUP_DIR"
TIMESTAMP=$(date -u +"%Y%m%d_%H%M%S")
DUMP_FILE="$BACKUP_DIR/hms_dump_$TIMESTAMP.sql.gz"
ENC_FILE="$BACKUP_DIR/hms_backup_$TIMESTAMP.enc"
MANIFEST_FILE="$BACKUP_DIR/hms_backup_$TIMESTAMP.manifest.json"

if [ -z "${DATABASE_URL:-}" ]; then
  echo "Error: DATABASE_URL environment variable is required." >&2
  exit 1
fi

echo "Starting HMS ET Database Backup: $TIMESTAMP"
pg_dump "$DATABASE_URL" --format=custom --no-owner --no-privileges --file="$DUMP_FILE"

PRE_HASH=$(sha256sum "$DUMP_FILE" | awk '{print $1}')
SIZE_BYTES=$(wc -c < "$DUMP_FILE")

if [ -n "$ENCRYPTION_KEY" ]; then
  echo "Encrypting backup with AES-256-CBC..."
  export ENC_PASS="$ENCRYPTION_KEY"
  openssl enc -aes-256-cbc -salt -pbkdf2 -iter 100000 -in "$DUMP_FILE" -out "$ENC_FILE" -pass env:ENC_PASS
  rm -f "$DUMP_FILE"
  FINAL_FILE="$ENC_FILE"
  ENCRYPTED=true
  ALGO="aes-256-cbc-pbkdf2"
else
  echo "No BACKUP_ENCRYPTION_KEY set. Storing unencrypted dump."
  FINAL_FILE="$DUMP_FILE"
  ENCRYPTED=false
  ALGO="none"
fi

POST_HASH=$(sha256sum "$FINAL_FILE" | awk '{print $1}')
FINAL_SIZE=$(wc -c < "$FINAL_FILE")

cat <<EOF > "$MANIFEST_FILE"
{
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "backup_file": "$(basename "$FINAL_FILE")",
  "encrypted": $ENCRYPTED,
  "algorithm": "$ALGO",
  "sha256": "$POST_HASH",
  "raw_sha256": "$PRE_HASH",
  "size_bytes": $FINAL_SIZE,
  "retention_days": $RETENTION_DAYS
}
EOF

echo "Backup complete: $FINAL_FILE (SHA-256: $POST_HASH)"

# Prune older than retention days
find "$BACKUP_DIR" -type f \( -name "hms_backup_*" -o -name "hms_dump_*" \) -mtime "+$RETENTION_DAYS" -delete
echo "Retention pruning applied for backups older than $RETENTION_DAYS days."
