# Card Town CSV Card Sheet Integrity & Schema Linter
# Validates CSV structure, headers, and core gameplay enum values.
# Usage: powershell -ExecutionPolicy Bypass -File lint_card_sheets.ps1

param(
    [string]$RepoRoot = (Resolve-Path "$PSScriptRoot\..\..").Path,
    [string[]]$Files = @()
)

$cardsDir = Join-Path $RepoRoot "data\cards"

Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "   CARD TOWN LINTER: CARD SHEET CSV INTEGRITY & SCHEMA    " -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "Repo Root: $RepoRoot"
Write-Host "Cards Directory: $cardsDir`n"

$hasErrors = $false

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
            $hasErrors = $true
        }
    }
} else {
    $csvFiles = Get-ChildItem -Path $cardsDir -Filter "*.csv" -ErrorAction SilentlyContinue
}

if (-not $csvFiles -or $csvFiles.Count -eq 0) {
    Write-Host "  [WARN] No CSV card sheets found to audit." -ForegroundColor Yellow
    exit 0
}

Write-Host "  Found $($csvFiles.Count) CSV sheet(s) to audit." -ForegroundColor Gray

# Canonical legal enums and headers (FR-DAT-006, FR-VAL-003)
$legalCardTypes   = @("UNIT", "STRUCTURE", "HEADQUARTERS", "MANEUVER")
$legalFactions    = @("CONCORDIA", "SOULLESS", "PATH", "ANIMUS", "NEUTRAL")
$legalRarities    = @("COMMON", "UNCOMMON", "RARE", "MYTHIC", "HQ")
$legalZones       = @("RESERVES", "MAIN_THEATER", "HEADQUARTERS", "MAIN_THEATER, RESERVES")
$legalBooleans    = @("true", "false")
$expectedAssetHeaders    = @("name","cardType","faction","subtypes","requisitionCost","tacticsCost","deploymentZones","rulesText","power","health","requisitionYield","tacticsYield","rarity","currentProxy","needsRework")
$expectedManeuverHeaders = @("name","cardType","faction","requisitionCost","tacticsCost","rulesText","rarity","currentProxy","needsRework")

foreach ($csv in $csvFiles) {
    try {
        $rawLines = Get-Content $csv.FullName
        if ($rawLines.Count -eq 0) {
            Write-Host "  [FAIL] $($csv.Name): File is empty!" -ForegroundColor Red
            $hasErrors = $true
            continue
        }

        # Check for duplicated headers on raw line 1
        $headerParts = @($rawLines[0].Split(',') | ForEach-Object { $_.Trim().Trim('"') })
        $headerCounts = @($headerParts | Group-Object | Where-Object { $_.Count -gt 1 })
        if ($headerCounts.Count -gt 0) {
            Write-Host "  [FAIL] $($csv.Name): Duplicate column headers found: $(($headerCounts.Name) -join ', ')" -ForegroundColor Red
            $hasErrors = $true
            continue
        }

        # Check for corrupted 'xz' column
        if ($headerParts -contains "xz") {
            Write-Host "  [FAIL] $($csv.Name): Corrupted column header 'xz' detected!" -ForegroundColor Red
            $hasErrors = $true
            continue
        }

        # Enforce exact camelCase header schema (15-column asset sheets or 9-column maneuver sheets)
        $expectedHeaders = if ($csv.Name -like "maneuvers-*.csv") { $expectedManeuverHeaders } else { $expectedAssetHeaders }
        if (($headerParts -join ',') -cne ($expectedHeaders -join ',')) {
            Write-Host "  [HEADER ERROR] $($csv.Name): Headers do not match canonical camelCase schema." -ForegroundColor Red
            Write-Host "    Expected: $($expectedHeaders -join ', ')" -ForegroundColor Yellow
            Write-Host "    Found:    $($headerParts -join ', ')" -ForegroundColor Yellow
            $hasErrors = $true
            continue
        }

        # Test Import-Csv
        $imported = @(Import-Csv -Path $csv.FullName -ErrorAction Stop)

        $csvErrors = 0
        for ($i = 0; $i -lt $imported.Count; $i++) {
            $r = $imported[$i]
            $rowNum = $i + 2

            # Skip empty placeholder rows awaiting final design
            if ([string]::IsNullOrWhiteSpace($r.name)) {
                continue
            }

            if ($r.cardType -and $legalCardTypes -notcontains $r.cardType) {
                Write-Host "  [ENUM ERROR] $($csv.Name):$rowNum -> Invalid cardType '$($r.cardType)'" -ForegroundColor Red
                $csvErrors++
            }
            if ($r.faction -and $legalFactions -notcontains $r.faction) {
                Write-Host "  [ENUM ERROR] $($csv.Name):$rowNum -> Invalid faction '$($r.faction)'" -ForegroundColor Red
                $csvErrors++
            }
            if ($r.rarity -and $legalRarities -notcontains $r.rarity) {
                Write-Host "  [ENUM ERROR] $($csv.Name):$rowNum -> Invalid rarity '$($r.rarity)'" -ForegroundColor Red
                $csvErrors++
            }
            if ($r.PSObject.Properties['deploymentZones'] -and $r.deploymentZones) {
                if ($legalZones -notcontains $r.deploymentZones) {
                    Write-Host "  [ENUM ERROR] $($csv.Name):$rowNum -> Invalid deploymentZones '$($r.deploymentZones)'" -ForegroundColor Red
                    $csvErrors++
                }
            }
            if ($legalBooleans -cnotcontains $r.currentProxy) {
                Write-Host "  [BOOL ERROR] $($csv.Name):$rowNum -> Invalid currentProxy '$($r.currentProxy)' (must be 'true' or 'false')" -ForegroundColor Red
                $csvErrors++
            }
            if ($legalBooleans -cnotcontains $r.needsRework) {
                Write-Host "  [BOOL ERROR] $($csv.Name):$rowNum -> Invalid needsRework '$($r.needsRework)' (must be 'true' or 'false')" -ForegroundColor Red
                $csvErrors++
            }
        }

        if ($csvErrors -eq 0) {
            Write-Host "  [PASS] $($csv.Name) ($($imported.Count) rows, $($headerParts.Count) columns)" -ForegroundColor Green
        } else {
            $hasErrors = $true
        }
    } catch {
        Write-Host "  [FAIL] $($csv.Name): Import error: $_" -ForegroundColor Red
        $hasErrors = $true
    }
}

Write-Host ""
if (-not $hasErrors) {
    Write-Host "  [PASS] All card sheet CSVs parsed cleanly with valid schemas!" -ForegroundColor Green
    Write-Host "==========================================================" -ForegroundColor Cyan
    exit 0
} else {
    Write-Host "  [FAIL] Card sheet CSV audit failed. Resolve highlighted errors above." -ForegroundColor Red
    Write-Host "==========================================================" -ForegroundColor Cyan
    exit 1
}
