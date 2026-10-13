# Card Town Unified Linter & Orchestrator
# Dispatches to modular linters:
#   - lint_frd.ps1         (FRD requirement IDs & cross-references)
#   - lint_card_sheets.ps1 (CSV parsing, headers & schema)
#   - lint_card_syntax.ps1 (Card rules text terminology & syntax)
#
# Usage:
#   powershell -ExecutionPolicy Bypass -File lint_cardtown.ps1              # Smart-runs only on changed files (Git status)
#   powershell -ExecutionPolicy Bypass -File lint_cardtown.ps1 -FRD         # Audit FRD references only
#   powershell -ExecutionPolicy Bypass -File lint_cardtown.ps1 -Cards       # Audit CSV sheet integrity & syntax only
#   powershell -ExecutionPolicy Bypass -File lint_cardtown.ps1 -Sheets      # Audit CSV sheet integrity only
#   powershell -ExecutionPolicy Bypass -File lint_cardtown.ps1 -Syntax      # Audit card syntax only
#   powershell -ExecutionPolicy Bypass -File lint_cardtown.ps1 -All         # Force full audit across all files

param(
    [string]$RepoRoot = (Resolve-Path "$PSScriptRoot\..\..").Path,
    [switch]$FRD,
    [switch]$Sheets,
    [switch]$Syntax,
    [switch]$Cards,
    [switch]$All,
    [switch]$ChangedOnly
)

$scriptDir = $PSScriptRoot

Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "               CARD TOWN LINTER DISPATCHER                " -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "Repo Root: $RepoRoot`n"

$runFrd = $false
$runSheets = $false
$runSyntax = $false
$changedFrds = @()
$changedCsvs = @()

# Explicit flags
if ($All) {
    $runFrd = $true
    $runSheets = $true
    $runSyntax = $true
} elseif ($FRD -or $Sheets -or $Syntax -or $Cards) {
    if ($FRD) { $runFrd = $true }
    if ($Sheets -or $Cards) { $runSheets = $true }
    if ($Syntax -or $Cards) { $runSyntax = $true }
} else {
    # Default mode: Smart change detection via Git
    Write-Host "[Detecting Changed Files via Git...]" -ForegroundColor Yellow
    try {
        $gitStatus = & git -C $RepoRoot status --porcelain 2>$null
        if ($LASTEXITCODE -eq 0 -and $gitStatus) {
            foreach ($line in $gitStatus) {
                # Format: XY <file path> or XY <file1> -> <file2>
                $relPath = ($line.Substring(3)).Trim()
                if ($relPath -match '->\s*(.+)$') {
                    $relPath = $matches[1].Trim()
                }

                if ($relPath -match '^docs[/\\]frd[/\\].+\.md$') {
                    $changedFrds += $relPath
                } elseif ($relPath -match '^data[/\\]cards[/\\].+\.csv$') {
                    $changedCsvs += $relPath
                }
            }

            if ($changedFrds.Count -gt 0) {
                Write-Host "  Detected $($changedFrds.Count) modified/added FRD document(s)." -ForegroundColor Gray
                $runFrd = $true
            }
            if ($changedCsvs.Count -gt 0) {
                Write-Host "  Detected $($changedCsvs.Count) modified/added Card CSV sheet(s)." -ForegroundColor Gray
                $runSheets = $true
                $runSyntax = $true
            }

            if (-not $runFrd -and -not $runSheets) {
                Write-Host "  No changes detected in docs/frd/ or data/cards/." -ForegroundColor Green
                Write-Host "  Nothing to lint! (Use -All to force a full audit across all files)`n" -ForegroundColor Gray
                Write-Host "==========================================================" -ForegroundColor Cyan
                exit 0
            }
        } else {
            Write-Host "  No uncommitted changes detected. (Use -All to force full audit across all files)`n" -ForegroundColor Gray
            Write-Host "==========================================================" -ForegroundColor Cyan
            exit 0
        }
    } catch {
        Write-Host "  Git change detection unavailable. Running all linters..." -ForegroundColor Yellow
        $runFrd = $true
        $runSheets = $true
        $runSyntax = $true
    }
}

$anyFailed = $false

# 1. FRD Linter
if ($runFrd) {
    Write-Host "`n>>> Running FRD Requirements Linter..." -ForegroundColor Magenta
    $frdScript = Join-Path $scriptDir "lint_frd.ps1"
    & powershell -ExecutionPolicy Bypass -File $frdScript -RepoRoot $RepoRoot
    if ($LASTEXITCODE -ne 0) { $anyFailed = $true }
}

# 2. Card Sheet CSV Integrity Linter
if ($runSheets) {
    Write-Host "`n>>> Running Card Sheet CSV Integrity Linter..." -ForegroundColor Magenta
    $sheetsScript = Join-Path $scriptDir "lint_card_sheets.ps1"
    if ($changedCsvs.Count -gt 0 -and -not $All) {
        & powershell -ExecutionPolicy Bypass -File $sheetsScript -RepoRoot $RepoRoot -Files $changedCsvs
    } else {
        & powershell -ExecutionPolicy Bypass -File $sheetsScript -RepoRoot $RepoRoot
    }
    if ($LASTEXITCODE -ne 0) { $anyFailed = $true }
}

# 3. Card Rules Syntax & Terminology Linter
if ($runSyntax) {
    Write-Host "`n>>> Running Card Rules Syntax & Terminology Linter..." -ForegroundColor Magenta
    $syntaxScript = Join-Path $scriptDir "lint_card_syntax.ps1"
    if ($changedCsvs.Count -gt 0 -and -not $All) {
        & powershell -ExecutionPolicy Bypass -File $syntaxScript -RepoRoot $RepoRoot -Files $changedCsvs
    } else {
        & powershell -ExecutionPolicy Bypass -File $syntaxScript -RepoRoot $RepoRoot
    }
    if ($LASTEXITCODE -ne 0) { $anyFailed = $true }
}

Write-Host "`n==========================================================" -ForegroundColor Cyan
if ($anyFailed) {
    Write-Host "  LINTER SUMMARY: One or more checks FAILED!             " -ForegroundColor Red
    Write-Host "==========================================================" -ForegroundColor Cyan
    exit 1
} else {
    Write-Host "  LINTER SUMMARY: All executed checks PASSED!             " -ForegroundColor Green
    Write-Host "==========================================================" -ForegroundColor Cyan
    exit 0
}
