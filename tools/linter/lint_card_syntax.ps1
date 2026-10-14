# Card Town Card Rules Text Syntax & Terminology Linter
# Audits card text in CSV sheets for canonical engine terminology and MTG-style shorthand.
# Usage: powershell -ExecutionPolicy Bypass -File lint_card_syntax.ps1

param(
    [string]$RepoRoot = (Resolve-Path "$PSScriptRoot\..\..").Path,
    [string[]]$Files = @(),
    [switch]$TreatWarningsAsErrors = $true
)

$cardsDir = Join-Path $RepoRoot "data\cards"

Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "   CARD TOWN LINTER: CARD RULES SYNTAX & TERMINOLOGY      " -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "Repo Root: $RepoRoot"
Write-Host "Cards Directory: $cardsDir`n"

# Determine which CSV files to audit
if ($Files.Count -gt 0) {
    $rawFiles = @()
    foreach ($entry in $Files) {
        $rawFiles += ($entry -split '[,;\r\n]')
    }
    $csvFiles = @()
    foreach ($f in $rawFiles) {
        $trimmed = $f.Trim()
        if (-not $trimmed) { continue }
        $p = if ([System.IO.Path]::IsPathRooted($trimmed)) { $trimmed } else { Join-Path $RepoRoot $trimmed }
        if (Test-Path $p) {
            $csvFiles += Get-Item $p
        } else {
            Write-Host "  [ERROR] Specified file not found: $trimmed" -ForegroundColor Red
        }
    }
} else {
    $csvFiles = Get-ChildItem -Path $cardsDir -Filter "*.csv" -ErrorAction SilentlyContinue
}

if (-not $csvFiles -or $csvFiles.Count -eq 0) {
    Write-Host "  [WARN] No CSV card sheets found to audit." -ForegroundColor Yellow
    exit 0
}

Write-Host "  Auditing rules text across $($csvFiles.Count) CSV sheet(s)..." -ForegroundColor Gray

$forbiddenPatterns = @(
    @{ Pattern = '\bfriendly\b'; Message = "Use 'allied' instead of 'friendly'" },
    @{ Pattern = '~\b'; Message = "Raw '~' MTG shorthand detected" },
    @{ Pattern = '\+\d+\s+attack\b'; Message = "Use '+N power' instead of '+N attack'" }
)

$terminologyIssues = 0

foreach ($csv in $csvFiles) {
    try {
        $imported = Import-Csv -Path $csv.FullName -ErrorAction Stop
        for ($i = 0; $i -lt $imported.Count; $i++) {
            $row = $imported[$i]
            $cardName = if ($row.name) { $row.name } elseif ($row.Name) { $row.Name } else { "Row $($i + 2)" }
            $text = if ($row.rulesText) { $row.rulesText } else { $row.Text }
            if (-not [string]::IsNullOrWhiteSpace($text)) {
                foreach ($rule in $forbiddenPatterns) {
                    if ($text -match $rule.Pattern) {
                        Write-Host "  [SYNTAX ISSUE] $($csv.Name) -> $cardName`: $($rule.Message) (Text: `"$text`")" -ForegroundColor Yellow
                        $terminologyIssues++
                    }
                }
            }
        }
    } catch {
        Write-Host "  [ERROR] $($csv.Name): Could not read for syntax check: $_" -ForegroundColor Red
        $terminologyIssues++
    }
}

Write-Host ""
if ($terminologyIssues -eq 0) {
    Write-Host "  [PASS] All card rules text conforms to standard engine terminology!" -ForegroundColor Green
    Write-Host "==========================================================" -ForegroundColor Cyan
    exit 0
} else {
    Write-Host "  [WARN] Found $terminologyIssues terminology / syntax issue(s)." -ForegroundColor Yellow
    Write-Host "==========================================================" -ForegroundColor Cyan
    if ($TreatWarningsAsErrors) {
        exit 1
    } else {
        exit 0
    }
}
