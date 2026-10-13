# Card Town FRD & Card Sheet Validation Linter
# Usage: powershell -ExecutionPolicy Bypass -File lint_cardtown.ps1

param(
    [string]$RepoRoot = (Resolve-Path "$PSScriptRoot\..\..").Path
)

$frdDir = Join-Path $RepoRoot "docs\frd"
$cardsDir = Join-Path $RepoRoot "data\cards"

Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "  CARD TOWN LINTER: FRD REQUIREMENTS & CARD SHEET AUDIT   " -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "Repo Root: $RepoRoot"
Write-Host "FRD Directory: $frdDir"
Write-Host "Cards Directory: $cardsDir`n"

$hasErrors = $false

# -----------------------------------------------------------------------------
# 1. AUDIT FRD SPECIFICATIONS & REQUIREMENT CROSS-REFERENCES
# -----------------------------------------------------------------------------
Write-Host "[1/3] Auditing FRD Requirement References..." -ForegroundColor Yellow

$frdFiles = @(
    "Game_Rules_and_Card_Modeling_FRD.md",
    "Client_Engine_Protocol_Networking_FRD.md",
    "Card_Data_Pipeline_FRD.md"
)

$reqDefs = @{}
$reqRefs = @()

foreach ($f in $frdFiles) {
    $filePath = Join-Path $frdDir $f
    if (-not (Test-Path $filePath)) {
        Write-Host "  [ERROR] Missing expected FRD file: $f" -ForegroundColor Red
        $hasErrors = $true
        continue
    }

    $lines = Get-Content $filePath
    for ($i = 0; $i -lt $lines.Count; $i++) {
        $line = $lines[$i]
        $matches = [regex]::Matches($line, 'FR-[A-Z]{3}-\d{3}[a-z]?')
        foreach ($m in $matches) {
            $val = $m.Value
            # Match header / list definition patterns
            if ($line -match '^(####|###|\*\*FR-|- \*\*FR-|\| \*\*FR-)') {
                if (-not $reqDefs.ContainsKey($val)) {
                    $reqDefs[$val] = @{ File = $f; Line = ($i + 1); Text = $line.Trim() }
                }
            }
            $reqRefs += [PSCustomObject]@{ Req = $val; File = $f; Line = ($i + 1); Text = $line.Trim() }
        }
    }
}

Write-Host "  Discovered $($reqDefs.Count) defined requirement IDs." -ForegroundColor Gray

# Check for dangling references
$uniqueRefs = $reqRefs | Select-Object -ExpandProperty Req -Unique | Sort-Object
$dangling = @()
foreach ($r in $uniqueRefs) {
    if (-not $reqDefs.ContainsKey($r)) {
        $dangling += $r
        $hasErrors = $true
        Write-Host "  [DANGLING REFERENCE] $r" -ForegroundColor Red
        $reqRefs | Where-Object { $_.Req -eq $r } | ForEach-Object {
            Write-Host "    -> Referenced in $($_.File):$($_.Line)" -ForegroundColor DarkRed
        }
    }
}

if ($dangling.Count -eq 0) {
    Write-Host "  [PASS] Zero dangling references! All requirement cross-references resolve cleanly." -ForegroundColor Green
} else {
    Write-Host "  [FAIL] Found $($dangling.Count) dangling references." -ForegroundColor Red
}

# -----------------------------------------------------------------------------
# 2. AUDIT CSV CARD SHEETS (HEADERS & SCHEMA)
# -----------------------------------------------------------------------------
Write-Host "`n[2/3] Auditing Card Sheet CSV Parsing & Headers..." -ForegroundColor Yellow

$csvFiles = Get-ChildItem -Path $cardsDir -Filter "*.csv"
Write-Host "  Found $($csvFiles.Count) CSV sheets to audit." -ForegroundColor Gray

foreach ($csv in $csvFiles) {
    try {
        $rawLines = Get-Content $csv.FullName
        if ($rawLines.Count -eq 0) {
            Write-Host "  [FAIL] $($csv.Name): File is empty!" -ForegroundColor Red
            $hasErrors = $true
            continue
        }

        # Check for duplicated headers on raw line 1
        $headerParts = $rawLines[0].Split(',') | ForEach-Object { $_.Trim().Trim('"') }
        $headerCounts = $headerParts | Group-Object | Where-Object { $_.Count -gt 1 }
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

        # Test Import-Csv
        $imported = Import-Csv -Path $csv.FullName -ErrorAction Stop
        
        # Validate target enums & required fields (FR-DAT-006, FR-VAL-003)
        $legalCardTypes = @("UNIT", "STRUCTURE", "HEADQUARTERS", "MANEUVER")
        $legalFactions = @("CONCORDIA", "SOULLESS", "PATH", "ANIMUS", "NEUTRAL")
        $legalRarities = @("COMMON", "UNCOMMON", "RARE", "MYTHIC", "HQ")
        $legalZones = @("RESERVES", "MAIN_THEATER", "HEADQUARTERS", "MAIN_THEATER, RESERVES")

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

# -----------------------------------------------------------------------------
# 3. AUDIT CARD RULES TEXT SYNTAX & TERMINOLOGY
# -----------------------------------------------------------------------------
Write-Host "`n[3/3] Auditing Card Text Terminology & Syntax Rules..." -ForegroundColor Yellow

$forbiddenPatterns = @(
    @{ Pattern = '\bfriendly\b'; Message = "Use 'allied' instead of 'friendly'" },
    @{ Pattern = '~\b'; Message = "Raw '~' MTG shorthand detected" },
    @{ Pattern = '\+\d+\s+attack\b'; Message = "Use '+N power' instead of '+N attack'" }
)

$terminologyIssues = 0

foreach ($csv in $csvFiles) {
    try {
        $imported = Import-Csv -Path $csv.FullName
        for ($i = 0; $i -lt $imported.Count; $i++) {
            $row = $imported[$i]
            $cardName = if ($row.name) { $row.name } elseif ($row.Name) { $row.Name } else { "Row $($i+1)" }
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
    } catch {}
}

if ($terminologyIssues -eq 0) {
    Write-Host "  [PASS] All card rules text conforms to standard engine terminology!" -ForegroundColor Green
} else {
    Write-Host "  [WARN] Found $terminologyIssues terminology / syntax warnings." -ForegroundColor Yellow
}

# -----------------------------------------------------------------------------
# SUMMARY
# -----------------------------------------------------------------------------
Write-Host "`n==========================================================" -ForegroundColor Cyan
if (-not $hasErrors) {
    Write-Host "  LINTER PASSED: All specifications & card sheets healthy!  " -ForegroundColor Green
} else {
    Write-Host "  LINTER FAILED: Resolve highlighted errors above.         " -ForegroundColor Red
}
Write-Host "==========================================================" -ForegroundColor Cyan
