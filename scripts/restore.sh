#!/usr/bin/env bash
set -euo pipefail
: "${RESTORE_DATABASE_URL:?Set an explicit empty target database URL}"
exec node "$(dirname "$0")/database-backup.mjs" restore "${1:?Backup file required}"
