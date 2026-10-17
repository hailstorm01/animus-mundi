<#
.SYNOPSIS
    Live capability and state snapshot for the docs workflow (feeds the docs-concierge persona).

.DESCRIPTION
    Discovers everything at run time, so it cannot go stale:
      skills - every .agents/skills/*/SKILL.md (name + start of its description)
      tools  - every tools/**/*.ps1 (synopsis + current parameter syntax)
      state  - git branch/unpushed commits, uncommitted docs/data changes, pre-commit hook install
               status, and pending audits (via tools/audit/audit_status.ps1)
    Read-only. Always exits 0.

.PARAMETER Section
    all (default), skills, tools, or state.

.EXAMPLE
    powershell -ExecutionPolicy Bypass -File tools/docs/whats_up.ps1 -Section state
#>

[CmdletBinding()]
param(
    [ValidateSet('all', 'skills', 'tools', 'state')]
    [string]$Section = 'all'
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$RepoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..\..")).Path

function Limit-Text([string]$Text, [int]$Max) {
    if ($Text.Length -le $Max) { return $Text }
    $cut = $Text.Substring(0, $Max)
    $space = $cut.LastIndexOf(' ')
    if ($space -gt 0) { $cut = $cut.Substring(0, $space) }
    return "$cut ..."
}

function Get-Git {
    # Never throws; returns output lines (empty on failure).
    param([Parameter(ValueFromRemainingArguments = $true)][string[]]$GitArgs)
    $prev = $ErrorActionPreference
    $ErrorActionPreference = "Continue"
    try { return @(& git -C $RepoRoot @GitArgs 2>$null) } finally { $ErrorActionPreference = $prev }
}

function Show-Skills {
    Write-Host "== SKILLS / PERSONAS (.agents/skills) ==" -ForegroundColor Cyan
    $files = @(Get-ChildItem (Join-Path $RepoRoot ".agents\skills") -Filter "SKILL.md" -Recurse -ErrorAction SilentlyContinue | Sort-Object FullName)
    foreach ($f in $files) {
        $text = [System.IO.File]::ReadAllText($f.FullName, [System.Text.Encoding]::UTF8)
        $front = [regex]::Match($text, '(?s)^---\r?\n(.*?)\r?\n---')
        if (-not $front.Success) { continue }
        $fm = $front.Groups[1].Value
        $nameM = [regex]::Match($fm, '(?m)^name:\s*(.+)$')
        $name = if ($nameM.Success) { $nameM.Groups[1].Value.Trim() } else { Split-Path (Split-Path $f.FullName) -Leaf }
        $desc = ""
        $block = [regex]::Match($fm, '(?ms)^description:\s*[>|]-?[ \t]*\r?\n((?:[ \t]+[^\r\n]*(?:\r?\n|$))+)')
        if ($block.Success) {
            $desc = (($block.Groups[1].Value -split '\r?\n') | ForEach-Object { $_.Trim() } | Where-Object { $_ }) -join ' '
        } else {
            $one = [regex]::Match($fm, '(?m)^description:\s*(.+)$')
            if ($one.Success) { $desc = $one.Groups[1].Value.Trim() }
        }
        $rel = $f.FullName.Substring($RepoRoot.Length + 1).Replace('\', '/')
        Write-Host "- $name ($rel)"
        Write-Host ("    " + (Limit-Text $desc 220))
    }
    Write-Host ""
}

function Show-Tools {
    Write-Host "== TOOLS (tools/**/*.ps1) ==" -ForegroundColor Cyan
    $toolsDir = Join-Path $RepoRoot "tools"
    $files = @(Get-ChildItem $toolsDir -Filter "*.ps1" -Recurse -ErrorAction SilentlyContinue | Sort-Object FullName)
    foreach ($f in $files) {
        $rel = $f.FullName.Substring($RepoRoot.Length + 1).Replace('\', '/')
        $text = [System.IO.File]::ReadAllText($f.FullName, [System.Text.Encoding]::UTF8)
        $syn = [regex]::Match($text, '(?ms)\.SYNOPSIS\s*\r?\n\s*(\S[^\r\n]*)')
        $summary = ""
        if ($syn.Success) {
            $summary = $syn.Groups[1].Value.Trim()
        } else {
            foreach ($line in ($text -split '\r?\n' | Select-Object -First 8)) {
                if ($line -match '^#(?!!)\s*(\S.*)$') { $summary = $Matches[1].Trim(); break }
            }
        }
        if (-not $summary) { $summary = "(no description)" }
        $syntax = ""
        try {
            $syntax = ((Get-Command $f.FullName -Syntax -ErrorAction Stop) -join ' ').Replace($f.FullName, $rel)
            $syntax = ($syntax -replace '\s*\[<CommonParameters>\]', '' -replace '\s+', ' ').Trim()
        } catch { $syntax = $rel }
        Write-Host "- $rel : $(Limit-Text $summary 140)"
        Write-Host ("    " + (Limit-Text $syntax 260))
    }
    Write-Host ""
}

function Show-State {
    Write-Host "== LIVE STATE ==" -ForegroundColor Cyan

    $branch = (Get-Git rev-parse --abbrev-ref HEAD | Select-Object -First 1)
    $ahead = (Get-Git rev-list --count '@{u}..HEAD' | Select-Object -First 1)
    if (-not $ahead) { $ahead = "?" }
    Write-Host "Git: branch '$branch', $ahead unpushed commit(s)"

    $changed = @(Get-Git status --porcelain -- docs data/cards tools .agents)
    if ($changed.Count -eq 0) {
        Write-Host "Uncommitted changes (docs/, data/cards/, tools/, .agents/): none"
    } else {
        Write-Host "Uncommitted changes (docs/, data/cards/, tools/, .agents/): $($changed.Count)"
        $changed | Select-Object -First 8 | ForEach-Object { Write-Host "    $_" }
        if ($changed.Count -gt 8) { Write-Host "    ... and $($changed.Count - 8) more" }
    }

    $hookSrc = Join-Path $RepoRoot "tools\linter\pre-commit"
    $hookDst = Join-Path $RepoRoot ".git\hooks\pre-commit"
    if (-not (Test-Path $hookDst)) {
        Write-Host "Pre-commit hook: NOT INSTALLED (run tools/linter/install_hooks.ps1)" -ForegroundColor Yellow
    } elseif ((Test-Path $hookSrc) -and ((Get-FileHash $hookSrc).Hash -ne (Get-FileHash $hookDst).Hash)) {
        Write-Host "Pre-commit hook: installed but OUT OF DATE (rerun tools/linter/install_hooks.ps1)" -ForegroundColor Yellow
    } else {
        Write-Host "Pre-commit hook: installed and current"
    }

    Write-Host ""
    Write-Host "Pending audits:"
    $auditScript = Join-Path $RepoRoot "tools\audit\audit_status.ps1"
    if (Test-Path $auditScript) {
        & powershell -NoProfile -ExecutionPolicy Bypass -File $auditScript
    } else {
        Write-Host "  (tools/audit/audit_status.ps1 not found)"
    }

    # Portfolio sync staleness check (PRODUCT_SUMMARY.md vs FRDs)
    $sumFile = Join-Path $RepoRoot "docs\PRODUCT_SUMMARY.md"
    if (Test-Path $sumFile) {
        $sumTimeRaw = (Get-Git log -1 --format=%ct -- "docs/PRODUCT_SUMMARY.md" | Select-Object -First 1)
        if ($sumTimeRaw -and ($sumTimeRaw -match '^\d+$')) {
            $sumTime = [int64]$sumTimeRaw
            $frdDir = Join-Path $RepoRoot "docs\frd"
            $staleFrds = @()
            if (Test-Path $frdDir) {
                $frdFiles = Get-ChildItem -Path $frdDir -Filter "*.md" -ErrorAction SilentlyContinue
                foreach ($f in $frdFiles) {
                    $relFrd = "docs/frd/" + $f.Name
                    $frdTimeRaw = (Get-Git log -1 --format=%ct -- $relFrd | Select-Object -First 1)
                    if ($frdTimeRaw -and ($frdTimeRaw -match '^\d+$')) {
                        if ([int64]$frdTimeRaw -gt $sumTime) {
                            $staleFrds += $f.Name
                        }
                    }
                }
            }
            if ($staleFrds.Count -gt 0) {
                Write-Host "Portfolio roadmap (docs/PRODUCT_SUMMARY.md): OUT OF SYNC" -ForegroundColor Yellow
                Write-Host "  Newer commits detected in: $($staleFrds -join ', ')"
                Write-Host "  (To synchronize, prompt @pm-lead to reconcile with the updated FRD)" -ForegroundColor Gray
            } else {
                Write-Host "Portfolio roadmap (docs/PRODUCT_SUMMARY.md): in sync with FRDs"
            }
        }
    }
    Write-Host ""
}

if ($Section -in 'all', 'skills') { Show-Skills }
if ($Section -in 'all', 'tools')  { Show-Tools }
if ($Section -in 'all', 'state')  { Show-State }
