# Database backup and restore

The repository contains tested manual database backup/restore tooling. Scheduling, off-host storage, production recovery targets and an approved retention policy still need deployment configuration. A database dump does not include private attachment bytes; take a coordinated encrypted filesystem snapshot of `HMS_ATTACHMENT_DIR` and retain its manifest with the database backup. Do not claim a complete hospital recovery drill until those files and the application are verified together.

## Setup

1. Install Node.js and PostgreSQL client tools matching the server major version. Set `PG_BIN` to their directory if they are not on PATH (Windows example: `C:/Program Files/PostgreSQL/18/bin`).
2. Generate a random 32-byte key with `node -e "console.log(require('node:crypto').randomBytes(32).toString('base64'))"`. Store it in a password manager or secret store, separately from backups. Supply it through `BACKUP_ENCRYPTION_KEY`. The example environment deliberately leaves this blank.
3. Supply `DATABASE_URL` privately. Set `BACKUP_DIR` and optionally `BACKUP_RETENTION_DAYS` (default 30; this is a configurable engineering default, not a legal retention recommendation).
4. Run `node scripts/database-backup.mjs backup <directory>`, or `scripts/backup.ps1` / `scripts/backup.sh`. Both wrappers use the same implementation.
5. Store the resulting `.enc` and `.enc.json` together. Version 2 uses AES-256-GCM, a fresh nonce, authentication tag and ciphertext/plaintext SHA-256 checks. The manifest is not independently signed; authenticated decryption detects ciphertext/tag modification. Plaintext staging uses a private temporary file and is removed after execution. Protect the OS temporary volume and backup-directory ACLs.
6. Configure the scheduler/off-host copy separately, monitor failures, and test key recovery. Retention prunes only paired files created by this format after a new backup succeeds. Older CBC-format files are not accepted by this tool; restore those only through a separately reviewed legacy procedure.

## Restore

1. Create a new empty database on an isolated host. Set `RESTORE_DATABASE_URL` explicitly; there is no fallback to the normal database URL.
2. Supply the matching backup key and run `node scripts/database-backup.mjs restore <backup.enc>`.
3. The tool verifies the manifest, authenticates decryption, checks the plaintext digest, and rejects a nonempty target. It uses `pg_restore --exit-on-error --single-transaction`; it never uses `--clean` and never converts errors to success.
4. Compare table counts, retained audit records, source identifiers and financial sums with the source export. Run the reconciliation tool and application QA, and restore/verify the separately backed-up private attachment directory.
5. A restore does not switch application traffic. Cutover requires its own verified migration and rollback procedure.

## Executed evidence

`node scripts/verify-backup-restore.mjs` uses two generated loopback-only databases and synthetic patient/audit records. On 2026-10-05 it passed encrypted backup, exact data restoration, wrong-key denial, tampering denial and nonempty-target denial, then removed only its generated databases. This is a tooling regression drill, not evidence of production RPO/RTO, live-data migration or off-host disaster recovery.

Run it with private `DATABASE_URL` and `PG_BIN` configured; it generates its own temporary encryption key. The normal development database is never restored or cleared.
