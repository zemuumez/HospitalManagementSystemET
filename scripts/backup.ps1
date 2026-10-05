param([string]$BackupDir="backups")
$ErrorActionPreference="Stop"
& node (Join-Path $PSScriptRoot 'database-backup.mjs') backup $BackupDir
if($LASTEXITCODE -ne 0){throw 'Backup failed'}
