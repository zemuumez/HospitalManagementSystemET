param([Parameter(Mandatory=$true)][string]$BackupFile)
$ErrorActionPreference="Stop"
if(-not $env:RESTORE_DATABASE_URL){throw 'Set RESTORE_DATABASE_URL to an empty target database'}
& node (Join-Path $PSScriptRoot 'database-backup.mjs') restore $BackupFile
if($LASTEXITCODE -ne 0){throw 'Restore failed'}
