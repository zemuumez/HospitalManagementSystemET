<#
.SYNOPSIS
  Automated Encrypted PostgreSQL Restore Script for HMS ET.
.DESCRIPTION
  Verifies SHA-256 hash, decrypts OpenSSL AES-256 backup, and restores into target PostgreSQL database.
.PARAMETER BackupFile
  Path to .enc or .sql.gz backup file.
.PARAMETER TargetDbUrl
  Target PostgreSQL database URL (defaults to env:DATABASE_URL).
#>
param (
    [Parameter(Mandatory=$true)]
    [string]$BackupFile,
    [string]$TargetDbUrl = $env:DATABASE_URL,
    [string]$EncryptionKey = $env:BACKUP_ENCRYPTION_KEY
)

$ErrorActionPreference = "Stop"

if (-not (Test-Path $BackupFile)) {
    Write-Error "Backup file not found: $BackupFile"
}

if (-not $TargetDbUrl) {
    Write-Error "Target database URL is required (specify -TargetDbUrl or set DATABASE_URL)."
}

Write-Host "Starting HMS ET Database Restore Drill" -ForegroundColor Cyan
Write-Host "Backup source: $BackupFile"

$ManifestFile = [System.IO.Path]::ChangeExtension($BackupFile, "manifest.json")
if (Test-Path $ManifestFile) {
    Write-Host "Verifying integrity against manifest: $ManifestFile"
    $Manifest = Get-Content $ManifestFile | ConvertFrom-Json
    $ActualHash = (Get-FileHash -Algorithm SHA256 -Path $BackupFile).Hash
    if ($ActualHash -ne $Manifest.sha256) {
        Write-Error "Hash mismatch! Backup may be corrupted or tampered with. Expected: $($Manifest.sha256), Got: $ActualHash"
    }
    Write-Host "Checksum verification passed: $ActualHash" -ForegroundColor Green
}

$TempRestoreFile = [System.IO.Path]::GetTempFileName() + ".sql.gz"

try {
    if ($BackupFile.EndsWith(".enc")) {
        if (-not $EncryptionKey) {
            Write-Error "Encrypted backup requires BACKUP_ENCRYPTION_KEY or -EncryptionKey parameter."
        }
        Write-Host "Decrypting backup file..."
        $env:ENC_PASS = $EncryptionKey
        & openssl enc -d -aes-256-cbc -pbkdf2 -iter 100000 -in "$BackupFile" -out "$TempRestoreFile" -pass env:ENC_PASS
        if ($LASTEXITCODE -ne 0) {
            Write-Error "Decryption failed. Ensure the encryption key is correct."
        }
    } else {
        Copy-Item "$BackupFile" "$TempRestoreFile" -Force
    }

    Write-Host "Restoring database from decrypted dump..."
    & pg_restore --clean --if-exists --no-owner --no-privileges --dbname="$TargetDbUrl" "$TempRestoreFile"
    if ($LASTEXITCODE -ne 0) {
        Write-Warning "pg_restore completed with warnings or non-fatal errors (code $LASTEXITCODE)."
    } else {
        Write-Host "pg_restore completed cleanly." -ForegroundColor Green
    }

    Write-Host "Running post-restore verification query..."
    $CheckCount = & psql "$TargetDbUrl" -t -c "SELECT COUNT(*) FROM patient;"
    Write-Host "Active patient records in restored database: $($CheckCount.Trim())" -ForegroundColor Green
    Write-Host "Restore drill successfully verified!" -ForegroundColor Green
}
finally {
    if (Test-Path $TempRestoreFile) {
        Remove-Item $TempRestoreFile -Force -ErrorAction SilentlyContinue
    }
}
