# Automated Backup, Encryption, Retention, and Restore Drill Runbook

## 1. Objectives & RPO/RTO Targets

| Metric | Target | Verification Method |
| :--- | :--- | :--- |
| **Recovery Point Objective (RPO)** | $\le 1\text{ hour}$ | Hourly WAL archiving / daily full snapshots + continuous log shipping |
| **Recovery Time Objective (RTO)** | $\le 15\text{ minutes}$ | Automated decryption and parallel pg_restore test drill |
| **Data Confidentiality** | AES-256-CBC with PBKDF2 | Zero plaintext backups at rest; keys stored separately in KMS/Vault |
| **Integrity Assurance** | SHA-256 pre/post hashing | Tamper-proof JSON manifests verified before decryption/restore |

---

## 2. Automated Backup Execution

### 2.1 Daily Scheduled Backup Job
Backups run automatically via cron / systemd timer / Windows Scheduled Task.

**Windows (PowerShell):**
```powershell
$env:DATABASE_URL = "postgres://hms_admin:secret@localhost:5432/hms_prod?sslmode=disable"
$env:BACKUP_ENCRYPTION_KEY = "k3y_fr0m_s3cur3_v4ult"
.\scripts\backup.ps1 -BackupDir "D:\HMS_Backups" -RetentionDays 30
```

**Linux (Bash):**
```bash
export DATABASE_URL="postgres://hms_admin:secret@localhost:5432/hms_prod?sslmode=disable"
export BACKUP_ENCRYPTION_KEY="k3y_fr0m_s3cur3_v4ult"
./scripts/backup.sh
```

### 2.2 Retention Policy
- **Daily Backups**: Kept for 30 days locally and replicated to off-site cloud cold storage.
- **Monthly Backups**: Kept for 7 years for medical compliance and tax audit requirements.
- **Automatic Pruning**: Backups older than the retention threshold are automatically purged by the script after manifest generation.

---

## 3. Demonstrated Restore Drill Procedure

The disaster recovery team must execute this drill quarterly on an isolated staging instance.

### Step 1: Isolate the Drill Environment
Ensure `TARGET_DB_URL` points to the designated drill database (e.g. `hms_drill_restore`), never production:
```bash
createdb -U postgres hms_drill_restore
```

### Step 2: Run Restore Script with Checksum Verification
```powershell
.\scripts\restore.ps1 -BackupFile "backups\hms_backup_20261005_020000.enc" -TargetDbUrl "postgres://postgres:secret@localhost:5432/hms_drill_restore?sslmode=disable"
```

The script will:
1. Parse `hms_backup_20261005_020000.manifest.json`.
2. Compute the SHA-256 hash of the `.enc` file and abort immediately if any byte has changed.
3. Decrypt the dump into a temporary memory/disk buffer.
4. Run `pg_restore --clean --if-exists`.
5. Execute smoke tests verifying core table integrity (`patient`, `"user"`, `encounter`).

### Step 3: Run Post-Restore Reconciliation Drill
Run the automated reconciliation tool against the restored database:
```bash
go run scripts/reconcile_import.go -db "postgres://postgres:secret@localhost:5432/hms_drill_restore?sslmode=disable" -out drill_report.json
```
Verify `drill_report.json` contains:
```json
{
  "status": "PASS",
  "integrityRules": {
    "zero_orphan_appointments": { "passed": true, "violations": 0 },
    "zero_orphan_encounters": { "passed": true, "violations": 0 },
    "non_negative_invoices": { "passed": true, "violations": 0 }
  }
}
```

### Step 4: Cleanup & Drill Sign-off
```bash
dropdb -U postgres hms_drill_restore
```
Log the drill timestamp, operator ID, time taken (RTO achieved), and sign off in the compliance log.
