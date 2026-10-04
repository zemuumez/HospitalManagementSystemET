# Production Deployment & Secret Management Runbook

## 1. Environment Separation & Security Controls

| Environment | Purpose | Database | TLS / Hostname | Auth URL |
| :--- | :--- | :--- | :--- | :--- |
| **Development** | Local engineering | Local container / SQLite test | `http://localhost:8080` | `http://localhost:3000` |
| **Staging** | Mirror testing & UAT drills | Isolated PostgreSQL instance | `https://staging.hospital.et` | `https://staging.hospital.et` |
| **Production** | Live clinical hospital operations | Redundant clustered PostgreSQL | `https://app.hospital.et` | `https://app.hospital.et` |

### 1.1 Secret Management Rules
1. **Never Commit Secrets**: Any file containing real credentials (`.env`, `secrets/`, `.pem`) must be ignored by `.gitignore`.
2. **File Permissions**: Secret files must be restricted to root/service accounts (`chmod 600 /etc/hms/secrets/*` or Windows ACL limited to SYSTEM/Administrators).
3. **Secret Rotation**: Database passwords, Better-Auth keys, Stripe webhook secrets, and Twilio/SMTP tokens rotate every 90 days.

---

## 2. Zero-Downtime Deployment Workflow

### Pre-Deployment Checks
1. Ensure all CI tests passed (unit, integration, migration linting).
2. Validate staging release with automated smoke checks:
   ```bash
   curl -f https://staging.hospital.et/healthz
   curl -f https://staging.hospital.et/readyz
   ```
3. Take a pre-deployment database snapshot:
   ```powershell
   .\scripts\backup.ps1 -BackupDir "D:\HMS_Backups\pre_deploy"
   ```

### Step 1: Execute Pending Database Migrations
Migrations are strictly additive and backward-compatible (Zero breaking schema mutations):
```bash
for file in $(ls db/migrations/*.sql | sort); do
  psql "$DATABASE_URL" -f "$file"
done
```

### Step 2: Build and Stage New Docker Containers
```bash
docker compose -f deployment/docker-compose.prod.yml build api worker
```

### Step 3: Rolling Service Update
```bash
docker compose -f deployment/docker-compose.prod.yml up -d --no-deps --remove-orphans api worker
```

### Step 4: Health & Readiness Verification
```bash
# Verify API responds ready
curl -s http://127.0.0.1:8080/readyz | grep '"status":"ready"'

# Verify Prometheus metrics endpoint
curl -s http://127.0.0.1:8080/metrics | grep 'hms_app_info'
```

---

## 3. Rollback Procedure

If `/readyz` fails or p95 latency spikes $> 500\text{ ms}$:
1. Revert to previous image tag:
   ```bash
   docker compose -f deployment/docker-compose.prod.yml rollback
   ```
2. If schema rollback is required, execute down-migration script or restore the pre-deploy snapshot.
3. Notify the incident response team.
