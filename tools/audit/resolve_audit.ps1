<#
.SYNOPSIS
    Closes a pending PM Devil's Advocate audit in Card Town.

.DESCRIPTION
    1. Validates the registry row, lints the target FRD, and verifies nothing unrelated is staged.
    2. Deletes the audit file from docs/audit/pending/ and commits it together with the target FRD
       (path-limited commit: nothing else is swept in).
    3. Marks the registry row RESOLVED (timestamp, resolution commit, source) and commits it.
    4. Pushes ONLY when -Push is given. Agents must not pass -Push unless the human asked for it.

    Every git call is exit-code checked. If the resolution commit fails (e.g. the pre-commit hook
    rejects it), the pending file and index are restored and the script aborts BEFORE touching the registry.

.PARAMETER AuditFile
    Filename or path of the pending audit (e.g. "2026-10-13-Card_Compiler_FRD.md").

.PARAMETER Source
    Resolution source label (default "Agentic (pm-audit-reviewer)").

.PARAMETER CommitMessage
    Custom message for the FRD remediation commit.

.PARAMETER Push
    Run 'git push' at the end. Off by default.

.PARAMETER DryRun
    Validate and report only; change nothing.

.EXAMPLE
    powershell -ExecutionPolicy Bypass -File tools/audit/resolve_audit.ps1 -AuditFile "2026-10-13-Card_Compiler_FRD.md"
#>

[CmdletBinding()]
param(
    [Parameter(Mandatory = $true, Position = 0)]
    [string]$AuditFile,

    [string]$Source = "Agentic (pm-audit-reviewer)",

    [string]$CommitMessage = "",

    [switch]$Push,

    [switch]$DryRun
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$RepoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..\..")).Path
$auditFileName = Split-Path $AuditFile -Leaf
$pendingRel = "docs/audit/pending/$auditFileName"
$pendingFilePath = Join-Path $RepoRoot ($pendingRel -replace '/', '\')
$registryRel = "docs/audit/audit_registry.md"
$registryPath = Join-Path $RepoRoot ($registryRel -replace '/', '\')

function Invoke-Git {
    # Runs git in the repo root and throws on a non-zero exit code
    # (native exit codes do not trigger $ErrorActionPreference on their own).
    param([Parameter(ValueFromRemainingArguments = $true)][string[]]$GitArgs)
    $out = & git -C $RepoRoot @GitArgs
    if ($LASTEXITCODE -ne 0) {
        throw "git $($GitArgs -join ' ') failed with exit code $LASTEXITCODE."
    }
    return $out
}

function Test-Git {
    # Returns $true when git exits 0; never throws, discards output.
    param([Parameter(ValueFromRemainingArguments = $true)][string[]]$GitArgs)
    $prev = $ErrorActionPreference
    $ErrorActionPreference = "Continue"
    try { & git -C $RepoRoot @GitArgs 2>&1 | Out-Null } finally { $ErrorActionPreference = $prev }
    return ($LASTEXITCODE -eq 0)
}

Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "  CARD TOWN: PM AUDIT RESOLUTION ENGINE" -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "Target Audit: $auditFileName"
Write-Host "Source:       $Source"
Write-Host ""

if (-not (Test-Path $registryPath)) { throw "Audit registry not found at '$registryPath'." }

# 1. Locate and validate the registry row
$registryContent = [System.IO.File]::ReadAllText($registryPath, [System.Text.Encoding]::UTF8)
$lines = $registryContent -split "`r?`n"
$targetLineIndex = -1
for ($i = 0; $i -lt $lines.Length; $i++) {
    if ($lines[$i].StartsWith("|") -and $lines[$i].Contains($auditFileName)) { $targetLineIndex = $i; break }
}
if ($targetLineIndex -eq -1) { throw "No registry row in '$registryRel' matches '$auditFileName'." }

$originalLine = $lines[$targetLineIndex]
$columns = @($originalLine.Split("|") | ForEach-Object { $_.Trim() })
if ($columns.Length -lt 9) { throw "Failed to parse registry row: $originalLine" }
if ($columns[5] -eq "RESOLVED") { throw "Registry row for '$auditFileName' is already RESOLVED." }

$targetDocRel = $columns[4].Trim('`')
if (-not $targetDocRel -or -not (Test-Path (Join-Path $RepoRoot $targetDocRel))) {
    throw "Target document '$targetDocRel' from the registry row does not exist."
}
Write-Host "[1/5] Registry row found (line $($targetLineIndex + 1)); target: $targetDocRel" -ForegroundColor Green

# 2. Guard: nothing unrelated may be staged (the commit below is path-limited anyway)
$allowed = @($targetDocRel, $pendingRel, $registryRel)
$unrelated = @(Invoke-Git diff --cached --name-only | Where-Object { $_ -and ($allowed -notcontains $_) })
if ($unrelated.Count -gt 0) {
    throw "Unrelated files are staged: $($unrelated -join ', '). Commit or unstage them first."
}

# 3. Lint BEFORE changing anything
& powershell -NoProfile -ExecutionPolicy Bypass -File (Join-Path $RepoRoot "tools\linter\lint_frd.ps1") -RepoRoot $RepoRoot -Files $targetDocRel
if ($LASTEXITCODE -ne 0) { throw "FRD linter failed for $targetDocRel. Fix dangling references before resolving." }
Write-Host "[2/5] FRD linter passed." -ForegroundColor Green

if ($DryRun) {
    Write-Host "[DRY-RUN] Would delete $pendingRel, commit it with $targetDocRel, mark the registry RESOLVED and commit." -ForegroundColor Yellow
    return
}

# 4. Delete the pending file (keep a backup for rollback), stage, and commit path-limited
$pendingExisted = Test-Path $pendingFilePath
$pendingBackup = if ($pendingExisted) { [System.IO.File]::ReadAllBytes($pendingFilePath) } else { $null }
$pendingTracked = Test-Git ls-files --error-unmatch '--' $pendingRel

$commitPaths = @($targetDocRel)
if ($pendingExisted) { Remove-Item -Path $pendingFilePath -Force }
if ($pendingTracked) { $commitPaths += $pendingRel }

$docBaseName = $auditFileName -replace '^\d{4}-\d{2}-\d{2}-', '' -replace '\.md$', ''
if ([string]::IsNullOrWhiteSpace($CommitMessage)) {
    $CommitMessage = "docs(frd): resolve PM devil's advocate audit for $docBaseName"
}

try {
    Invoke-Git add -A '--' @commitPaths | Out-Null
    $staged = @(Invoke-Git diff --cached --name-only '--' @commitPaths)
    if ($staged.Count -gt 0) {
        Invoke-Git commit -m $CommitMessage '--' @commitPaths | Out-Null
        $resolutionCommit = (Invoke-Git rev-parse --short HEAD | Select-Object -First 1).Trim()
        Write-Host "[3/5] Committed resolution ${resolutionCommit}: $($staged -join ', ')" -ForegroundColor Green
    } else {
        # FRD edits were already committed earlier; attribute the last commit that touched the FRD.
        $resolutionCommit = (Invoke-Git log -1 --format=%h '--' $targetDocRel | Select-Object -First 1).Trim()
        Write-Host "[3/5] Nothing new to commit; attributing last FRD commit $resolutionCommit" -ForegroundColor Yellow
    }
} catch {
    Write-Host "[ABORT] Resolution commit failed: $($_.Exception.Message)" -ForegroundColor Red
    Test-Git reset -q '--' @commitPaths | Out-Null
    if ($pendingExisted -and -not (Test-Path $pendingFilePath)) {
        [System.IO.File]::WriteAllBytes($pendingFilePath, $pendingBackup)
        Write-Host "        Restored $pendingRel; registry untouched." -ForegroundColor Yellow
    }
    throw
}

# 5. Update the registry row and commit it (path-limited)
$timestamp = (Get-Date).ToString("yyyy-MM-dd HH:mm:ss")
$newRow = "| $($columns[1]) | $($columns[2]) | ``$auditFileName`` | $($columns[4]) | RESOLVED | $timestamp | ``$resolutionCommit`` | $Source |"
$lines[$targetLineIndex] = $newRow
[System.IO.File]::WriteAllText($registryPath, ($lines -join "`r`n"), [System.Text.Encoding]::UTF8)
Write-Host "[4/5] Registry updated: $newRow" -ForegroundColor Green

Invoke-Git add '--' $registryRel | Out-Null
Invoke-Git commit -m "docs(audit): log resolution of $auditFileName in audit_registry.md" '--' $registryRel | Out-Null
$registryCommit = (Invoke-Git rev-parse --short HEAD | Select-Object -First 1).Trim()

if ($Push) {
    Invoke-Git push | Out-Null
    Write-Host "[5/5] Pushed to remote." -ForegroundColor Green
} else {
    Write-Host "[5/5] Not pushed (local commits only; pass -Push to publish)." -ForegroundColor Yellow
}

Write-Host ""
Write-Host "AUDIT RESOLVED: $auditFileName | Resolution: $resolutionCommit | Registry: $registryCommit" -ForegroundColor Cyan
