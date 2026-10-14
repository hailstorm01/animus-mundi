<#
.SYNOPSIS
    Read-only status report for pending Card Town audits: staleness and severity counts.

.DESCRIPTION
    For each file in docs/audit/pending/ this reports:
      - the registry row (audited commit, target FRD, status)
      - STALE if the target FRD has commits after the audited commit (commit-based, not date-based)
      - counts of Blocker / High / Medium / Low findings, and whether a human is required
        (any Blocker or High finding, or a non-empty Socratic Interrogation Register)
    Exit code is always 0. Used at reviewer intake and by the --clear-backlog loop.

.EXAMPLE
    powershell -ExecutionPolicy Bypass -File tools/audit/audit_status.ps1
#>

[CmdletBinding()]
param()

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$RepoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..\..")).Path
$pendingDir = Join-Path $RepoRoot "docs\audit\pending"
$registryPath = Join-Path $RepoRoot "docs\audit\audit_registry.md"

$registryRows = @()
if (Test-Path $registryPath) {
    $registryRows = @([System.IO.File]::ReadAllLines($registryPath, [System.Text.Encoding]::UTF8) | Where-Object { $_.StartsWith("|") })
}

$files = @()
if (Test-Path $pendingDir) { $files = @(Get-ChildItem $pendingDir -Filter "*.md" | Sort-Object Name) }

if ($files.Count -eq 0) {
    Write-Host "No pending audits." -ForegroundColor Green
    return
}

foreach ($f in $files) {
    $row = $registryRows | Where-Object { $_.Contains($f.Name) } | Select-Object -First 1
    $stale = "NO REGISTRY ROW"
    $detail = ""
    $target = "?"
    $hash = "?"
    if ($row) {
        $cols = @($row.Split("|") | ForEach-Object { $_.Trim() })
        $hash = $cols[2].Trim('`')
        $target = $cols[4].Trim('`')
        & git -C $RepoRoot cat-file -e "$hash^{commit}" 2>$null
        if ($LASTEXITCODE -ne 0) {
            $stale = "UNKNOWN (commit $hash not found)"
        } else {
            $newer = @(& git -C $RepoRoot log --format="%h %s" "$hash..HEAD" -- $target)
            if ($newer.Count -gt 0) {
                $stale = "STALE"
                $detail = ($newer | Select-Object -First 5) -join "; "
            } else {
                $stale = "CURRENT"
            }
        }
    }

    $counts = [ordered]@{ Blocker = 0; High = 0; Medium = 0; Low = 0 }
    $section = $null
    $socratic = 0
    $inSocratic = $false
    foreach ($line in [System.IO.File]::ReadAllLines($f.FullName, [System.Text.Encoding]::UTF8)) {
        if ($line -match '^##\s+\d+\.\s+Socratic') { $inSocratic = $true; $section = $null; continue }
        if ($line -match '^##\s' -and $line -notmatch 'Socratic') { $inSocratic = $false }
        if ($line -match '^###\s') {
            $section = $null
            foreach ($k in $counts.Keys) { if ($line -match $k) { $section = $k } }
            if ($line -match 'Critical') { $section = 'Blocker' }
            continue
        }
        if ($section -and $line -match '^\s*-\s+\*\*\[DA-[A-Z]+-\d+\]') { $counts[$section]++ }
        if ($inSocratic -and $line -match '^\s*\d+\.\s') { $socratic++ }
    }

    $human = ($counts.Blocker + $counts.High) -gt 0 -or $socratic -gt 0
    Write-Host ""
    Write-Host $f.Name -ForegroundColor Cyan
    Write-Host "  Target: $target | Audited at: $hash | Freshness: $stale"
    if ($detail) { Write-Host "  FRD commits since audit: $detail" -ForegroundColor Yellow }
    Write-Host "  Findings: Blocker=$($counts.Blocker) High=$($counts.High) Medium=$($counts.Medium) Low=$($counts.Low) | Socratic questions=$socratic"
    Write-Host ("  Human required: " + $(if ($human) { "YES (Blocker/High findings or open Socratic questions)" } else { "NO (loop-eligible)" }))
}
