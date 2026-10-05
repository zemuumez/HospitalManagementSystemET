#!/usr/bin/env bash
set -euo pipefail
exec node "$(dirname "$0")/database-backup.mjs" backup "${BACKUP_DIR:-backups}"
