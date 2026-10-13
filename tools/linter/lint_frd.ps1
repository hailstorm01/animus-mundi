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
    "Game_Rules_and_Card_Modeling_FRD.md",
    "Client_Engine_Protocol_Networking_FRD.md",
    "Card_Data_Pipeline_FRD.md",
    "Card_Compiler_FRD.md"
)

# Collect all FRD files to inspect
if ($Files.Count -gt 0) {
    $targetFiles = $Files | ForEach-Object {
        if ([System.IO.Path]::IsPathRooted($_)) { $_ } else { Join-Path $RepoRoot $_ }
    }
} else {
    $existingFrds = Get-ChildItem -Path $frdDir -Filter "*.md" -ErrorAction SilentlyContinue | Select-Object -ExpandProperty FullName
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

$reqDefs = @{}
$reqRefs = @()

foreach ($filePath in $targetFiles) {
    if (-not (Test-Path $filePath)) {
        Write-Host "  [ERROR] File not found: $filePath" -ForegroundColor Red
        $hasErrors = $true
        continue
    }

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
            $reqRefs += [PSCustomObject]@{ Req = $val; File = $fileName; Line = ($i + 1); Text = $line.Trim() }
        }
    }
}

Write-Host "  Discovered $($reqDefs.Count) defined requirement IDs across $($targetFiles.Count) FRD files." -ForegroundColor Gray

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
