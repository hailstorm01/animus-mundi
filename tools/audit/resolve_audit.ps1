<#
.SYNOPSIS
    Resolves a pending PM Devil's Advocate audit in Card Town.

.DESCRIPTION
    Deletes the audit file from docs/audit/pending/, commits any pending FRD remediations
    and the file removal, captures the resolution commit hash, and updates docs/audit/audit_registry.md
    with the resolution timestamp, commit hash, and source attribution. Finally, commits the registry
    update and pushes to version control.

.PARAMETER AuditFile
    The filename or relative path of the pending audit (e.g. "2026-10-13-Card_Compiler_FRD.md").

.PARAMETER Source
    The source attribution for the resolution (defaults to "Agentic (pm-audit-reviewer)").

.PARAMETER CommitMessage
    Custom git commit message for the FRD remediation commit.

.PARAMETER SkipPush
    If specified, skips running 'git push'.

.PARAMETER DryRun
    If specified, simulates the actions without modifying files or git commits.

.EXAMPLE
    powershell -ExecutionPolicy Bypass -File tools/audit/resolve_audit.ps1 -AuditFile "2026-10-13-Card_Compiler_FRD.md"
#>

[CmdletBinding()]
param(
    [Parameter(Mandatory = $true, Position = 0)]
    [string]$AuditFile,

    [string]$Source = "Agentic (pm-audit-reviewer)",

    [string]$CommitMessage = "",

    [switch]$SkipPush,

    [switch]$DryRun
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$RepoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..\..")).Path
$auditFileName = Split-Path $AuditFile -Leaf

Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "  CARD TOWN: PM AUDIT RESOLUTION ENGINE                  " -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "Repository Root:  $RepoRoot"
Write-Host "Target Audit:     $auditFileName"
Write-Host "Source:           $Source"
Write-Host ""

$pendingDir = Join-Path $RepoRoot "docs\audit\pending"
$pendingFilePath = Join-Path $pendingDir $auditFileName
$registryPath = Join-Path $RepoRoot "docs\audit\audit_registry.md"

if (-not (Test-Path $registryPath)) {
    throw "Audit registry not found at '$registryPath'."
}

# 1. Locate entry in audit_registry.md
$registryContent = [System.IO.File]::ReadAllText($registryPath, [System.Text.Encoding]::UTF8)
$lines = $registryContent -split "`r?`n"
$targetLineIndex = -1
$originalLine = ""

for ($i = 0; $i -lt $lines.Length; $i++) {
    if ($lines[$i] -match [regex]::Escape($auditFileName)) {
        $targetLineIndex = $i
        $originalLine = $lines[$i]
        break
    }
}

if ($targetLineIndex -eq -1) {
    throw "Could not find an audit registry entry in '$registryPath' matching '$auditFileName'."
}

Write-Host "[1/4] Found matching entry in audit_registry.md (Line $($targetLineIndex + 1))" -ForegroundColor Green
Write-Host "      Current: $originalLine" -ForegroundColor DarkGray

# Extract target document path from registry row
$rawCols = $originalLine.Split("|")
$targetDocRel = ""
if ($rawCols.Length -ge 5) {
    $targetDocRel = $rawCols[4].Trim().Trim('`')
}

if ($DryRun) {
    Write-Host "[DRY-RUN] Would delete: $pendingFilePath" -ForegroundColor Yellow
    Write-Host "[DRY-RUN] Would commit FRD changes ($targetDocRel) and file deletion" -ForegroundColor Yellow
    Write-Host "[DRY-RUN] Would update registry status to RESOLVED and commit" -ForegroundColor Yellow
    return
}

# 2. Delete pending audit file
if (Test-Path $pendingFilePath) {
    Remove-Item -Path $pendingFilePath -Force
    Write-Host "[2/4] Deleted pending audit file: docs/audit/pending/$auditFileName" -ForegroundColor Green
} else {
    Write-Host "[2/4] Note: $auditFileName was already removed from pending/." -ForegroundColor Yellow
}

# 3. Stage FRD edits and pending file removal, and commit
$docBaseName = $auditFileName -replace '^\d{4}-\d{2}-\d{2}-', '' -replace '\.md$', ''
if ([string]::IsNullOrWhiteSpace($CommitMessage)) {
    $CommitMessage = "docs(frd): resolve PM devil's advocate audit for $docBaseName"
}

# Add specifically targeted FRD changes (or fallback to docs/frd/) and deleted file
if ($targetDocRel -and (Test-Path (Join-Path $RepoRoot $targetDocRel))) {
    git -C $RepoRoot add $targetDocRel "docs/audit/pending/$auditFileName"
} else {
    git -C $RepoRoot add "docs/frd/" "docs/audit/pending/$auditFileName"
}

# Check if there are staged changes
$stagedDiff = git -C $RepoRoot diff --cached --name-only
$resolutionCommit = ""

if ($stagedDiff) {
    Write-Host "      Staged files for resolution commit:" -ForegroundColor DarkGray
    $stagedDiff | ForEach-Object { Write-Host "        - $_" -ForegroundColor DarkGray }
    git -C $RepoRoot commit -m $CommitMessage
    $resolutionCommit = (git -C $RepoRoot rev-parse --short HEAD).Trim()
    Write-Host "[3/4] Committed resolution: $resolutionCommit ('$CommitMessage')" -ForegroundColor Green
} else {
    # If working tree was already committed, use current HEAD
    $resolutionCommit = (git -C $RepoRoot rev-parse --short HEAD).Trim()
    Write-Host "[3/4] No uncommitted FRD changes staged. Using current HEAD: $resolutionCommit" -ForegroundColor Yellow
}

# 4. Update audit_registry.md
$timestamp = (Get-Date).ToString("yyyy-MM-dd HH:mm:ss")

# Parse markdown table line:
# Format: | Date Generated | Commit Hash | Source Audit File | Target Document | Status | Resolution Date | Resolution Commit | Source of Resolution |
$rawCols = $originalLine.Split("|")
$columns = @()
foreach ($c in $rawCols) {
    $columns += $c.Trim()
}

if ($columns.Length -ge 9) {
    $cleanedSourceFile = "`"$auditFileName`"" -replace '"', '`'
    $commitFormatted = "`"$resolutionCommit`"" -replace '"', '`'
    $colDate = $columns[1]
    $colGenCommit = $columns[2]
    $colTargetDoc = $columns[4]

    $newRow = "| $colDate | $colGenCommit | $cleanedSourceFile | $colTargetDoc | RESOLVED | $timestamp | $commitFormatted | $Source |"
    $lines[$targetLineIndex] = $newRow
    $newRegistryContent = [string]::Join("`r`n", $lines)
    [System.IO.File]::WriteAllText($registryPath, $newRegistryContent, [System.Text.Encoding]::UTF8)
    Write-Host "[4/4] Updated audit_registry.md:" -ForegroundColor Green
    Write-Host "      $newRow" -ForegroundColor Cyan
} else {
    throw "Failed to parse registry row columns: $originalLine"
}

# Commit the registry update
git -C $RepoRoot add "docs/audit/audit_registry.md"
git -C $RepoRoot commit -m "docs(audit): log resolution of $auditFileName in audit_registry.md"
$registryCommit = (git -C $RepoRoot rev-parse --short HEAD).Trim()
Write-Host "      Committed registry update: $registryCommit" -ForegroundColor Green

# Push changes
if (-not $SkipPush) {
    Write-Host "Pushing changes to remote..." -ForegroundColor DarkGray
    git -C $RepoRoot push
    Write-Host "Push complete!" -ForegroundColor Green
} else {
    Write-Host "Skipping push (-SkipPush was specified)." -ForegroundColor Yellow
}

Write-Host ""
Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "  AUDIT RESOLUTION COMPLETE: $auditFileName" -ForegroundColor Cyan
Write-Host "  Resolution Commit: $resolutionCommit | Registry: $registryCommit" -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan
