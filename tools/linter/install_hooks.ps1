# Script to install Git pre-commit hooks for Card Town
param(
    [string]$RepoRoot = (Resolve-Path "$PSScriptRoot\..\..").Path
)

$hookSource = Join-Path $PSScriptRoot "pre-commit"
$hooksDir = Join-Path $RepoRoot ".git\hooks"
$hookDest = Join-Path $hooksDir "pre-commit"

if (-not (Test-Path $hooksDir)) {
    Write-Host "[ERROR] Git hooks directory not found at $hooksDir" -ForegroundColor Red
    exit 1
}

Copy-Item -Path $hookSource -Destination $hookDest -Force
Write-Host "[SUCCESS] Installed Card Town pre-commit hook into $hookDest" -ForegroundColor Green
