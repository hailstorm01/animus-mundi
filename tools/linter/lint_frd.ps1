# Card Town FRD Requirement Cross-Reference Linter
# Validates requirement IDs (FR-XXX-YYY) and checks for dangling references.
# Usage: powershell -ExecutionPolicy Bypass -File lint_frd.ps1

param(
    [string]$RepoRoot = (Resolve-Path "$PSScriptRoot\..\..").Path,
    [string[]]$Files = @()
)

$frdDir = Join-Path $RepoRoot "docs\frd"

Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "  CARD TOWN LINTER: FRD REQUIREMENTS & CROSS-REFERENCES   " -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "Repo Root: $RepoRoot"
Write-Host "FRD Directory: $frdDir`n"

$hasErrors = $false

$expectedFiles = @(
    "01_Game_Rules_and_Card_Modeling_FRD.md",
    "02_Card_Data_Pipeline_FRD.md",
    "03_Card_Compiler_FRD.md",
    "04_Client_Engine_Protocol_Networking_FRD.md"
)

$existingFrds = Get-ChildItem -Path $frdDir -Filter "*.md" -ErrorAction SilentlyContinue | Select-Object -ExpandProperty FullName

# Collect all FRD files to inspect
if ($Files.Count -gt 0) {
    $rawFiles = @()
    foreach ($entry in $Files) {
        $rawFiles += ($entry -split '[,;\r\n]')
    }
    $targetFiles = @()
    foreach ($f in $rawFiles) {
        $trimmed = $f.Trim()
        if (-not $trimmed) { continue }
        $p = if ([System.IO.Path]::IsPathRooted($trimmed)) { $trimmed } else { Join-Path $RepoRoot $trimmed }
        if (Test-Path $p) {
            $targetFiles += $p
        } else {
            Write-Host "  [ERROR] Specified file not found: $trimmed" -ForegroundColor Red
            $hasErrors = $true
        }
    }
} else {
    $targetFiles = $existingFrds
}

# Verify expected core FRD files exist
foreach ($ef in $expectedFiles) {
    $expectedPath = Join-Path $frdDir $ef
    if (-not (Test-Path $expectedPath)) {
        Write-Host "  [ERROR] Missing expected core FRD file: $ef" -ForegroundColor Red
        $hasErrors = $true
    }
}

# Step 1: Scan ALL existing FRDs to build the complete requirement definition catalog
$reqDefs = @{}
foreach ($filePath in $existingFrds) {
    $fileName = Split-Path $filePath -Leaf
    $lines = Get-Content $filePath
    for ($i = 0; $i -lt $lines.Count; $i++) {
        $line = $lines[$i]
        $matches = [regex]::Matches($line, 'FR-[A-Z]{3}-\d{3}[a-z]?')
        foreach ($m in $matches) {
            $val = $m.Value
            # Match header / list definition patterns
            if ($line -match '^(####|###|\*\*FR-|- \*\*FR-|\| \*\*FR-)') {
                if (-not $reqDefs.ContainsKey($val)) {
                    $reqDefs[$val] = @{ File = $fileName; Line = ($i + 1); Text = $line.Trim() }
                }
            }
        }
    }
}

Write-Host "  Discovered $($reqDefs.Count) defined requirement IDs across $($existingFrds.Count) FRD files." -ForegroundColor Gray

# Step 2: Scan target file(s) for references to validate against the master definitions
$reqRefs = @()
foreach ($filePath in $targetFiles) {
    $fileName = Split-Path $filePath -Leaf
    $lines = Get-Content $filePath
    for ($i = 0; $i -lt $lines.Count; $i++) {
        $line = $lines[$i]
        $matches = [regex]::Matches($line, 'FR-[A-Z]{3}-\d{3}[a-z]?')
        foreach ($m in $matches) {
            $reqRefs += [PSCustomObject]@{ Req = $m.Value; File = $fileName; Line = ($i + 1); Text = $line.Trim() }
        }
    }
}
Write-Host "  Checking $($reqRefs.Count) requirement references across $($targetFiles.Count) target FRD file(s)..." -ForegroundColor Gray

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

Write-Host ""
if ($dangling.Count -eq 0 -and -not $hasErrors) {
    Write-Host "  [PASS] Zero dangling references! All requirement cross-references resolve cleanly." -ForegroundColor Green
    Write-Host "==========================================================" -ForegroundColor Cyan
    exit 0
} else {
    Write-Host "  [FAIL] Found $($dangling.Count) dangling references or missing core FRD files." -ForegroundColor Red
    Write-Host "==========================================================" -ForegroundColor Cyan
    exit 1
}
