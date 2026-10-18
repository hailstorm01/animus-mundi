<#
.SYNOPSIS
    Slices an approved Card Town TDD (docs/tdd/*.md) into a GitHub Milestone and dependency-ordered Issues.

.DESCRIPTION
    Ingests a lint-clean Technical Design Document (TDD), parses its Section 6 Deliverable &
    Implementation Slice Manifest (### [SLICE-XX-YY]), topologically sorts the slice DAG, and
    either previews (-DryRun, default) or publishes (-Publish) the Milestone, Labels, and
    self-contained GitHub Issues via the GitHub CLI (gh).
    Idempotent: matches existing issues by '[SLICE-XX-YY]' title prefix so re-runs never
    create duplicates, and resolves upstream slice IDs to real GitHub Issue numbers ('Depends on #N').

.PARAMETER TddFile
    Path to the target TDD markdown file (e.g., docs/tdd/03_Card_Compiler_TDD.md).

.PARAMETER Repo
    Optional GitHub owner/repo override (auto-detected from git origin remote by default).

.PARAMETER DryRun
    Preview mode (default when -Publish is omitted). Prints Milestone, Labels, topological slice
    table, and rendered issue body preview without creating or modifying anything on GitHub.

.PARAMETER Publish
    Executes live creation/synchronization of the GitHub Milestone, Labels, and Slice Issues.

.PARAMETER UpdateExisting
    When combined with -Publish, updates the body, labels, and milestone of already-existing
    GitHub issues matching [SLICE-XX-YY] instead of skipping them.

.PARAMETER ShowBodies
    In DryRun mode, prints the full rendered Markdown body of every slice issue.

.EXAMPLE
    powershell -ExecutionPolicy Bypass -File tools/tdd/publish_tdd_issues.ps1 -TddFile "docs/tdd/03_Card_Compiler_TDD.md" -DryRun
    powershell -ExecutionPolicy Bypass -File tools/tdd/publish_tdd_issues.ps1 -TddFile "docs/tdd/03_Card_Compiler_TDD.md" -Publish
#>

[CmdletBinding()]
param(
    [Parameter(Mandatory = $true)]
    [string]$TddFile,
    [string]$Repo = "",
    [switch]$DryRun,
    [switch]$Publish,
    [switch]$UpdateExisting,
    [switch]$ShowBodies
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$RepoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..\..")).Path
$tddFullPath = if ([System.IO.Path]::IsPathRooted($TddFile)) { $TddFile } else { Join-Path $RepoRoot $TddFile }

if (-not (Test-Path $tddFullPath)) {
    Write-Host "[ERROR] TDD file not found: $TddFile" -ForegroundColor Red
    exit 1
}
$tddFullPath = (Resolve-Path $tddFullPath).Path
$tddRelPath = $tddFullPath.Substring($RepoRoot.Length + 1).Replace('\', '/')

# Default to DryRun unless -Publish is explicitly specified
$isLivePublish = [bool]$Publish
if (-not $isLivePublish) {
    $DryRun = $true
}

# Step 1: Mandatory Quality Gate via lint_tdd.ps1
$lintScript = Join-Path $RepoRoot "tools\linter\lint_tdd.ps1"
Write-Host ">>> Running mandatory pre-slice TDD linter gate..." -ForegroundColor Magenta
& powershell -NoProfile -ExecutionPolicy Bypass -File $lintScript -RepoRoot $RepoRoot -Files $tddFullPath
if ($LASTEXITCODE -ne 0) {
    Write-Host "[ABORT] TDD failed linting. Fix lint errors before slicing or publishing GitHub issues." -ForegroundColor Red
    exit 1
}

# Step 2: Resolve GitHub repository (owner/repo)
if (-not $Repo) {
    $remoteUrl = (& git -C $RepoRoot remote get-url origin 2>$null | Select-Object -First 1)
    if ($remoteUrl -and $remoteUrl -match 'github\.com[:/]([^/]+/[^/.]+?)(?:\.git)?$') {
        $Repo = $Matches[1]
    }
}
if (-not $Repo) {
    $Repo = "hailstorm01/animus-mundi"
}

# Step 3: Parse TDD Metadata & Deliverable Slices
$lines = [System.IO.File]::ReadAllLines($tddFullPath, [System.Text.Encoding]::UTF8)

$sourceFrdRel = ""
$frdCommitHash = ""
$tddVersion = ""
$ghMilestone = ""

foreach ($line in ($lines | Select-Object -First 40)) {
    if ($line -match '^\s*-\s+\*\*Source FRD:\*\*\s*(.+)$') {
        $rawFrd = $Matches[1].Trim()
        if ($rawFrd -match '(\d{2}_[A-Za-z0-9_]+_FRD\.md)') {
            $sourceFrdRel = "docs/frd/$($Matches[1])"
        } else {
            $sourceFrdRel = $rawFrd
        }
    }
    if ($line -match '^\s*-\s+\*\*FRD Commit Hash:\*\*\s*`?([^`\s]+)`?') { $frdCommitHash = $Matches[1].Trim() }
    if ($line -match '^\s*-\s+\*\*TDD Version:\*\*\s*`?([^`\r\n]+)`?') { $tddVersion = $Matches[1].Trim() }
    if ($line -match '^\s*-\s+\*\*GitHub Milestone:\*\*\s*`?([^`\r\n]+)`?') { $ghMilestone = $Matches[1].Trim() }
}

$slices = [ordered]@{}
$sliceOrder = New-Object System.Collections.Generic.List[string]
$currentSliceId = $null
$currentSubSection = $null

for ($i = 0; $i -lt $lines.Count; $i++) {
    $line = $lines[$i]
    if ($line -match '^###\s+\[(SLICE-\d{2}-\d{2})\]\s+(.+)$') {
        $currentSliceId = $Matches[1].Trim()
        $sliceTitle = $Matches[2].Trim()
        $currentSubSection = $null
        $sliceOrder.Add($currentSliceId)
        $slices[$currentSliceId] = @{
            Id                  = $currentSliceId
            Title               = $sliceTitle
            IssueTitle          = "[$currentSliceId] $sliceTitle"
            Labels              = @()
            DependsOnList       = @()
            RequirementsCovered = @()
            TddSections         = ""
            TargetFiles         = New-Object System.Collections.Generic.List[string]
            InTargetFiles       = $false
            BlueprintLines      = New-Object System.Collections.Generic.List[string]
            AcceptanceLines     = New-Object System.Collections.Generic.List[string]
            VerifyLines         = New-Object System.Collections.Generic.List[string]
        }
        continue
    }

    if ($currentSliceId -and $line -match '^##\s+[^#]') {
        $currentSliceId = $null
        $currentSubSection = $null
        continue
    }

    if ($currentSliceId) {
        $s = $slices[$currentSliceId]
        if ($line -match '^####\s+(.+)$') {
            $subTitle = $Matches[1].Trim()
            $s.InTargetFiles = $false
            if ($subTitle -match '(?i)^Implementation Blueprint') {
                $currentSubSection = 'blueprint'
            } elseif ($subTitle -match '(?i)^Acceptance Criteria') {
                $currentSubSection = 'acceptance'
            } elseif ($subTitle -match '(?i)^Verification Commands') {
                $currentSubSection = 'verify'
            } else {
                $currentSubSection = 'other'
            }
            continue
        }

        if (-not $currentSubSection) {
            if ($line -match '^\s*-\s+\*\*Labels:\*\*\s*(.+)$') {
                $rawLabels = $Matches[1] -split ','
                $parsedLabels = @()
                foreach ($rl in $rawLabels) {
                    $clean = $rl.Trim().Trim('`')
                    if ($clean) { $parsedLabels += $clean }
                }
                $s.Labels = $parsedLabels
                $s.InTargetFiles = $false
            } elseif ($line -match '^\s*-\s+\*\*Depends On:\*\*\s*(.+)$') {
                $depRaw = $Matches[1].Trim()
                if ($depRaw -notmatch '(?i)^`?None`?$') {
                    $s.DependsOnList = @([regex]::Matches($depRaw, 'SLICE-\d{2}-\d{2}') | ForEach-Object { $_.Value })
                }
                $s.InTargetFiles = $false
            } elseif ($line -match '^\s*-\s+\*\*Requirements Covered:\*\*\s*(.+)$') {
                $s.RequirementsCovered = @([regex]::Matches($Matches[1], 'FR-[A-Z]{3}-\d{3}[a-z]?') | ForEach-Object { $_.Value })
                $s.InTargetFiles = $false
            } elseif ($line -match '^\s*-\s+\*\*TDD Sections:\*\*\s*(.+)$') {
                $s.TddSections = $Matches[1].Trim()
                $s.InTargetFiles = $false
            } elseif ($line -match '^\s*-\s+\*\*Target Files:\*\*\s*(.*)$') {
                $s.InTargetFiles = $true
                $inlineFile = $Matches[1].Trim()
                if ($inlineFile) { $s.TargetFiles.Add($inlineFile) }
            } elseif ($s.InTargetFiles -and $line -match '^\s+-\s+(.+)$') {
                $s.TargetFiles.Add($Matches[1].Trim())
            } elseif ($line.Trim() -ne "") {
                $s.InTargetFiles = $false
            }
        } else {
            if ($currentSubSection -eq 'blueprint') {
                $s.BlueprintLines.Add($line)
            } elseif ($currentSubSection -eq 'acceptance') {
                $s.AcceptanceLines.Add($line)
            } elseif ($currentSubSection -eq 'verify') {
                $s.VerifyLines.Add($line)
            }
        }
    }
}

# Step 4: Topological Sort (Kahn's / DFS post-order) so upstream dependencies always precede downstream
$topoOrder = New-Object System.Collections.Generic.List[string]
$visited = @{}
foreach ($sid in $sliceOrder) { $visited[$sid] = $false }

function Add-TopoSlice([string]$NodeId) {
    if ($visited[$NodeId]) { return }
    $visited[$NodeId] = $true
    foreach ($dep in $slices[$NodeId].DependsOnList) {
        if ($slices.Contains($dep) -and -not $visited[$dep]) {
            Add-TopoSlice $dep
        }
    }
    $script:topoOrder.Add($NodeId)
}

foreach ($sid in $sliceOrder) {
    Add-TopoSlice $sid
}

# Collect all unique labels across all slices
$allLabels = New-Object System.Collections.Generic.List[string]
foreach ($sid in $topoOrder) {
    foreach ($lbl in $slices[$sid].Labels) {
        if (-not $allLabels.Contains($lbl)) { $allLabels.Add($lbl) }
    }
}

# Helper to format GitHub Issue Markdown body for a slice
function Format-SliceIssueBody([hashtable]$Slice, [hashtable]$IssueNumberMap) {
    $depText = "None *(Foundational Slice -- Ready to Start Immediately)*"
    if ($Slice.DependsOnList.Count -gt 0) {
        $depParts = @()
        foreach ($d in $Slice.DependsOnList) {
            if ($IssueNumberMap.ContainsKey($d)) {
                $depParts += "#$($IssueNumberMap[$d]) (``$d``)"
            } else {
                $depParts += "``$d``"
            }
        }
        $depText = $depParts -join ", "
    }

    $reqList = ($Slice.RequirementsCovered | ForEach-Object { "``$_``" }) -join ", "
    $targetFilesBlock = ($Slice.TargetFiles | ForEach-Object { "- $_" }) -join "`n"
    $blueprintBlock = ($Slice.BlueprintLines -join "`n").Trim()
    $acceptanceBlock = ($Slice.AcceptanceLines -join "`n").Trim()
    $verifyBlock = ($Slice.VerifyLines -join "`n").Trim()

    return @"
## Deliverable Context & Traceability

- **Slice ID:** ``$($Slice.Id)``
- **Source TDD:** [$tddRelPath]($tddRelPath) *(Version: ``$tddVersion``)*
- **Source FRD:** [$sourceFrdRel]($sourceFrdRel) *(Audited FRD Commit: ``$frdCommitHash``)*
- **TDD Reference Sections:** $($Slice.TddSections)
- **FRD Requirements Covered:** $reqList
- **Depends On:** $depText

---

## Target Files

$targetFilesBlock

---

## Implementation Blueprint & Signatures

$blueprintBlock

---

## Acceptance Criteria

$acceptanceBlock

---

## Verification Commands

$verifyBlock

---

> [!IMPORTANT]
> **Zero-Guesswork Execution Protocol (Engineer / Agent):**
> 1. Read the referenced **TDD Sections ($($Slice.TddSections))** in ``$tddRelPath`` before writing code.
> 2. Implement strictly the types, signatures, and target files scoped to ``$($Slice.Id)``. Do not invent unapproved schemas or alter FRD rules.
> 3. Run the **Verification Commands** above and ``tools/linter/lint_cardtown.ps1`` until all checks exit ``0``.
"@
}

# Step 5: Check existing GitHub Issues (read-only if authenticated) for idempotent mapping
$existingIssuesBySlice = @{}
$ghAvailable = $false
try {
    $null = & gh auth status 2>$null
    if ($LASTEXITCODE -eq 0) {
        $ghAvailable = $true
        $rawIssuesJson = (& gh issue list --repo $Repo --state all --limit 500 --json number,title,url,state 2>$null) -join "`n"
        if ($LASTEXITCODE -eq 0 -and $rawIssuesJson) {
            $parsedIssues = $rawIssuesJson | ConvertFrom-Json
            foreach ($iss in $parsedIssues) {
                if ($iss.title -match '^\[(SLICE-\d{2}-\d{2})\]') {
                    $sid = $Matches[1]
                    if (-not $existingIssuesBySlice.ContainsKey($sid)) {
                        $existingIssuesBySlice[$sid] = @{
                            Number = $iss.number
                            Title  = $iss.title
                            Url    = $iss.url
                            State  = $iss.state
                        }
                    }
                }
            }
        }
    }
} catch {
    $ghAvailable = $false
}

$issueNumberMap = @{}
foreach ($k in $existingIssuesBySlice.Keys) {
    $issueNumberMap[$k] = $existingIssuesBySlice[$k].Number
}

Write-Host ""
Write-Host "==========================================================" -ForegroundColor Cyan
if ($isLivePublish) {
    Write-Host "  TDD ISSUE SLICER: LIVE GITHUB PUBLICATION (-Publish)    " -ForegroundColor Cyan
} else {
    Write-Host "  TDD ISSUE SLICER: DRY-RUN PREVIEW (-DryRun)             " -ForegroundColor Cyan
}
Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "Repository      : $Repo"
Write-Host "Source TDD      : $tddRelPath ($tddVersion)"
Write-Host "Source FRD      : $sourceFrdRel (Commit: $frdCommitHash)"
Write-Host "GitHub Milestone: $ghMilestone"
Write-Host "Labels to Ensure: $($allLabels -join ', ')"
Write-Host "Total Slices    : $($topoOrder.Count) (Topologically Ordered)`n"

if (-not $isLivePublish) {
    Write-Host "== TOPOLOGICAL DELIVERABLE SLICE PLAN ==" -ForegroundColor Yellow
    $idx = 1
    foreach ($sid in $topoOrder) {
        $s = $slices[$sid]
        $deps = if ($s.DependsOnList.Count -gt 0) { $s.DependsOnList -join ', ' } else { "None (Ready)" }
        $reqs = $s.RequirementsCovered -join ', '
        $actionTag = if ($existingIssuesBySlice.ContainsKey($sid)) {
            if ($UpdateExisting) { "[UPDATE #$($existingIssuesBySlice[$sid].Number)]" } else { "[SKIP EXISTING #$($existingIssuesBySlice[$sid].Number)]" }
        } else {
            "[CREATE NEW]"
        }
        Write-Host ("  {0,2}. {1} {2}" -f $idx, $actionTag, $s.IssueTitle) -ForegroundColor Cyan
        Write-Host "      Depends On  : $deps"
        Write-Host "      Requirements: $reqs"
        Write-Host "      Target Files: $($s.TargetFiles.Count) file(s) ($($s.TargetFiles -join '; '))"
        $idx++
    }

    if ($ShowBodies -or $topoOrder.Count -gt 0) {
        $previewList = if ($ShowBodies) { $topoOrder } else { @($topoOrder[0]) }
        foreach ($psid in $previewList) {
            Write-Host "`n----------------------------------------------------------" -ForegroundColor DarkGray
            Write-Host "  SAMPLE ISSUE BODY PREVIEW: $($slices[$psid].IssueTitle)" -ForegroundColor Gray
            Write-Host "----------------------------------------------------------" -ForegroundColor DarkGray
            Write-Host (Format-SliceIssueBody $slices[$psid] $issueNumberMap)
        }
    }

    Write-Host "`n[DRY-RUN COMPLETE] No changes were made to GitHub ($Repo)." -ForegroundColor Green
    Write-Host "To publish live after human confirmation, run with -Publish." -ForegroundColor Gray
    exit 0
}

# Step 6: Live Publish Mode (-Publish)
if (-not $ghAvailable) {
    Write-Host "[ERROR] GitHub CLI (gh) is not authenticated or cannot reach GitHub. Run 'gh auth login' first." -ForegroundColor Red
    exit 1
}

# 6.1 Ensure GitHub Milestone exists
Write-Host ">>> Ensuring GitHub Milestone '$ghMilestone' exists in $Repo..." -ForegroundColor Magenta
$milestonesJson = (& gh api "repos/$Repo/milestones?state=all&per_page=100" 2>$null) -join "`n"
$milestoneExists = $false
if ($LASTEXITCODE -eq 0 -and $milestonesJson) {
    $milestones = $milestonesJson | ConvertFrom-Json
    foreach ($m in $milestones) {
        if ($m.title -eq $ghMilestone) {
            $milestoneExists = $true
            Write-Host "  [OK] Milestone already exists (#$($m.number): $ghMilestone)" -ForegroundColor Green
            break
        }
    }
}
if (-not $milestoneExists) {
    $desc = "Engineering deliverables sliced from $tddRelPath ($tddVersion) governing $sourceFrdRel."
    $null = & gh api "repos/$Repo/milestones" -X POST -f title="$ghMilestone" -f description="$desc" 2>$null
    if ($LASTEXITCODE -ne 0) {
        Write-Host "[ERROR] Failed to create GitHub Milestone '$ghMilestone' in $Repo." -ForegroundColor Red
        exit 1
    }
    Write-Host "  [CREATED] Milestone '$ghMilestone'" -ForegroundColor Green
}

# 6.2 Ensure Labels exist
Write-Host ">>> Ensuring Labels exist in $Repo..." -ForegroundColor Magenta
foreach ($lbl in $allLabels) {
    $color = if ($lbl -like "type:*") { "0E8A16" }
             elseif ($lbl -like "phase:*") { "1D76DB" }
             elseif ($lbl -like "component:*") { "5319E7" }
             elseif ($lbl -like "tdd:*") { "FBCA04" }
             else { "EDEDED" }
    $null = & gh label create $lbl --repo $Repo --color $color --force 2>$null
}
Write-Host "  [OK] Verified $($allLabels.Count) label(s): $($allLabels -join ', ')" -ForegroundColor Green

# 6.3 Publish or Update Slice Issues in Topological Order
Write-Host ">>> Publishing Deliverable Slice Issues in topological order..." -ForegroundColor Magenta
$results = @()

foreach ($sid in $topoOrder) {
    $s = $slices[$sid]
    $bodyText = Format-SliceIssueBody $s $issueNumberMap
    $tmpBodyFile = Join-Path ([System.IO.Path]::GetTempPath()) "tdd_issue_${sid}_$PID.md"

    try {
        [System.IO.File]::WriteAllText($tmpBodyFile, $bodyText, (New-Object System.Text.UTF8Encoding($false)))

        if ($existingIssuesBySlice.ContainsKey($sid)) {
            $existingNum = $existingIssuesBySlice[$sid].Number
            $existingUrl = $existingIssuesBySlice[$sid].Url
            if ($UpdateExisting) {
                $editArgs = @("issue", "edit", "$existingNum", "--repo", $Repo, "--title", $s.IssueTitle, "--milestone", $ghMilestone, "--body-file", $tmpBodyFile)
                foreach ($lbl in $s.Labels) { $editArgs += @("--add-label", $lbl) }
                $null = & gh @editArgs
                if ($LASTEXITCODE -ne 0) {
                    throw "gh issue edit failed for #$existingNum ($sid)"
                }
                Write-Host "  [UPDATED] #$existingNum : $($s.IssueTitle) -> $existingUrl" -ForegroundColor Yellow
                $results += [PSCustomObject]@{ Slice = $sid; Issue = "#$existingNum"; Action = "UPDATED"; Url = $existingUrl }
            } else {
                Write-Host "  [SKIPPED] #$existingNum : $($s.IssueTitle) (already exists -> $existingUrl)" -ForegroundColor Gray
                $results += [PSCustomObject]@{ Slice = $sid; Issue = "#$existingNum"; Action = "SKIPPED"; Url = $existingUrl }
            }
        } else {
            $createArgs = @("issue", "create", "--repo", $Repo, "--title", $s.IssueTitle, "--milestone", $ghMilestone, "--body-file", $tmpBodyFile)
            foreach ($lbl in $s.Labels) { $createArgs += @("--label", $lbl) }
            $createdOut = (& gh @createArgs 2>&1) -join "`n"
            if ($LASTEXITCODE -ne 0) {
                throw "gh issue create failed for ${sid}: $createdOut"
            }
            $createdUrl = ($createdOut -split '\r?\n' | Where-Object { $_ -match 'https://github\.com/.+/issues/\d+' } | Select-Object -Last 1).Trim()
            $newNum = "?"
            if ($createdUrl -match '/issues/(\d+)$') {
                $newNum = $Matches[1]
                $issueNumberMap[$sid] = $newNum
            }
            Write-Host "  [CREATED] #$newNum : $($s.IssueTitle) -> $createdUrl" -ForegroundColor Green
            $results += [PSCustomObject]@{ Slice = $sid; Issue = "#$newNum"; Action = "CREATED"; Url = $createdUrl }
        }
    } finally {
        if (Test-Path $tmpBodyFile) { Remove-Item $tmpBodyFile -Force -ErrorAction SilentlyContinue }
    }
}

Write-Host "`n==========================================================" -ForegroundColor Cyan
Write-Host "  GITHUB PUBLICATION SUMMARY ($($results.Count) Slices)" -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan
foreach ($r in $results) {
    Write-Host ("  {0,-12} | {1,-6} | {2,-8} | {3}" -f $r.Slice, $r.Issue, $r.Action, $r.Url)
}
Write-Host "==========================================================" -ForegroundColor Cyan
exit 0
