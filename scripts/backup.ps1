<#
.SYNOPSIS
  Automated Encrypted PostgreSQL Backup Script for HMS ET.
.DESCRIPTION
  Performs full pg_dump, generates SHA-256 checksum, encrypts using OpenSSL AES-256,
  and prunes backups older than retention days.
.PARAMETER BackupDir
  Target directory for backups.
.PARAMETER RetentionDays
  Number of days to keep daily backups.
#>
param (
    [string]$BackupDir = "backups",
    [int]$RetentionDays = 30,
    [string]$EncryptionKey = $env:BACKUP_ENCRYPTION_KEY
)

$ErrorActionPreference = "Stop"

if (-not (Test-Path $BackupDir)) {
    New-Item -ItemType Directory -Path $BackupDir -Force | Out-Null
}

$Timestamp = (Get-Date).ToString("yyyyMMdd_HHmmss")
$DumpFile = Join-Path $BackupDir "hms_dump_$Timestamp.sql.gz"
$EncryptedFile = Join-Path $BackupDir "hms_backup_$Timestamp.enc"
$ManifestFile = Join-Path $BackupDir "hms_backup_$Timestamp.manifest.json"

Write-Host "Starting HMS ET Database Backup: $Timestamp" -ForegroundColor Cyan

# 1. Execute pg_dump with gzip compression
if (-not $env:DATABASE_URL) {
    Write-Error "DATABASE_URL environment variable is required."
}

Write-Host "Dumping PostgreSQL database..."
& pg_dump "$env:DATABASE_URL" --format=custom --no-owner --no-privileges --file="$DumpFile"
if ($LASTEXITCODE -ne 0) {
    Write-Error "pg_dump failed with exit code $LASTEXITCODE"
}

# 2. Compute SHA-256 Checksum before encryption
$PreHash = (Get-FileHash -Algorithm SHA256 -Path $DumpFile).Hash
$SizeBytes = (Get-Item $DumpFile).Length

# 3. Encrypt using OpenSSL if encryption key is present, otherwise store secured file
if ($EncryptionKey) {
    Write-Host "Encrypting dump with AES-256-CBC..."
    $env:ENC_PASS = $EncryptionKey
    & openssl enc -aes-256-cbc -salt -pbkdf2 -iter 100000 -in "$DumpFile" -out "$EncryptedFile" -pass env:ENC_PASS
    Remove-Item $DumpFile -Force
    $FinalFile = $EncryptedFile
    $Encrypted = $true
} else {
    Write-Host "No BACKUP_ENCRYPTION_KEY set. Storing unencrypted compressed dump." -ForegroundColor Yellow
    $FinalFile = $DumpFile
    $Encrypted = $false
}

$PostHash = (Get-FileHash -Algorithm SHA256 -Path $FinalFile).Hash
$FinalSize = (Get-Item $FinalFile).Length

# 4. Generate Manifest
$Manifest = [PSCustomObject]@{
    timestamp = (Get-Date).ToUniversalTime().ToString("o")
    backup_file = (Split-Path $FinalFile -Leaf)
    encrypted = $Encrypted
    algorithm = if ($Encrypted) { "aes-256-cbc-pbkdf2" } else { "none" }
    sha256 = $PostHash
    raw_sha256 = $PreHash
    size_bytes = $FinalSize
    retention_days = $RetentionDays
}
$Manifest | ConvertTo-Json -Depth 4 | Out-File -FilePath $ManifestFile -Encoding utf8

Write-Host "Backup successfully created: $FinalFile" -ForegroundColor Green
Write-Host "Manifest saved: $ManifestFile (SHA-256: $PostHash)" -ForegroundColor Green

# 5. Prune backups older than RetentionDays
$Cutoff = (Get-Date).AddDays(-$RetentionDays)
Get-ChildItem -Path $BackupDir -Include "hms_backup_*", "hms_dump_*" | Where-Object { $_.CreationTime -lt $Cutoff } | ForEach-Object {
    Write-Host "Pruning expired backup: $($_.Name)" -ForegroundColor DarkGray
    Remove-Item $_.FullName -Force
}
