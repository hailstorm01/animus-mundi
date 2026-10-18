<#
.SYNOPSIS
    Card Town TDD Requirement Coverage, Cross-Reference & Slice DAG Linter

.DESCRIPTION
    Validates Technical Design Documents (TDDs) in docs/tdd/*.md:
      1. Header metadata (Source FRD link, FRD Commit Hash, TDD Version, GitHub Milestone)
      2. Zero dangling FR-XXX-YYY references against the master FRD catalog in docs/frd/
      3. 100% requirement coverage of the Source FRD (all Active requirements mapped in the
         Traceability Matrix and assigned to at least one Deliverable Slice; all Subsumed
         requirements accounted for)
      4. Deliverable Slice Manifest (### [SLICE-XX-YY]) schema completeness and acyclic
         dependency graph (DAG) validation for tools/tdd/publish_tdd_issues.ps1.

.PARAMETER RepoRoot
    Root path of the repository.

.PARAMETER Files
    Optional specific TDD markdown file(s) to lint. Defaults to all docs/tdd/*.md files.

.EXAMPLE
    powershell -ExecutionPolicy Bypass -File tools/linter/lint_tdd.ps1
    powershell -ExecutionPolicy Bypass -File tools/linter/lint_tdd.ps1 -Files "docs/tdd/03_Card_Compiler_TDD.md"
#>

param(
    [string]$RepoRoot = (Resolve-Path "$PSScriptRoot\..\..").Path,
    [string[]]$Files = @()
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$frdDir = Join-Path $RepoRoot "docs\frd"
$tddDir = Join-Path $RepoRoot "docs\tdd"

Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "  CARD TOWN LINTER: TDD COVERAGE, REFS & SLICE MANIFEST   " -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "Repo Root: $RepoRoot"
Write-Host "TDD Directory: $tddDir`n"

$hasErrors = $false

# Collect target TDD files
$targetFiles = @()
if ($Files.Count -gt 0) {
    $rawFiles = @()
    foreach ($entry in $Files) {
        $rawFiles += ($entry -split '[,;\r\n]')
    }
    foreach ($f in $rawFiles) {
        $trimmed = $f.Trim()
        if (-not $trimmed) { continue }
        $p = if ([System.IO.Path]::IsPathRooted($trimmed)) { $trimmed } else { Join-Path $RepoRoot $trimmed }
        if (Test-Path $p) {
            $targetFiles += (Resolve-Path $p).Path
        } else {
            Write-Host "  [ERROR] Specified TDD file not found: $trimmed" -ForegroundColor Red
            $hasErrors = $true
        }
    }
} else {
    if (Test-Path $tddDir) {
        $targetFiles = @(Get-ChildItem -Path $tddDir -Filter "*.md" -ErrorAction SilentlyContinue |
            Where-Object { $_.Name -ne "README.md" } |
            Sort-Object Name |
            Select-Object -ExpandProperty FullName)
    }
}

if ($targetFiles.Count -eq 0 -and -not $hasErrors) {
    Write-Host "  No TDD documents found in docs/tdd/ yet. Skipping TDD lint." -ForegroundColor Gray
    Write-Host "==========================================================" -ForegroundColor Cyan
    exit 0
}

# Step 1: Scan all FRDs in docs/frd/ to build master requirement catalog & per-FRD ownership
$existingFrds = @(Get-ChildItem -Path $frdDir -Filter "*.md" -ErrorAction SilentlyContinue | Sort-Object Name)
$allReqDefs = @{}
$frdOwnedReqs = @{}
$subsumedReqs = @{}

foreach ($frdFile in $existingFrds) {
    $fn = $frdFile.Name
    $frdOwnedReqs[$fn] = New-Object System.Collections.Generic.List[string]
    $lines = [System.IO.File]::ReadAllLines($frdFile.FullName, [System.Text.Encoding]::UTF8)

    for ($i = 0; $i -lt $lines.Count; $i++) {
        $line = $lines[$i]

        # 1a. General master catalog matching (same as lint_frd.ps1)
        if ($line -match '^(####|###|\*\*FR-|- \*\*FR-|\| \*\*FR-)') {
            foreach ($m in [regex]::Matches($line, 'FR-[A-Z]{3}-\d{3}[a-z]?')) {
                $val = $m.Value
                if (-not $allReqDefs.ContainsKey($val)) {
                    $allReqDefs[$val] = @{ File = $fn; Line = ($i + 1) }
                }
            }
        }

        # 1b. Canonical per-FRD ownership & Active vs Subsumed classification
        $ownedId = $null
        $isSub = $false
        if ($line -match '^(?:####|###)\s+(FR-[A-Z]{3}-\d{3}[a-z]?)') {
            $ownedId = $Matches[1]
            $next1 = if ($i + 1 -lt $lines.Count) { $lines[$i + 1] } else { "" }
            $next2 = if ($i + 2 -lt $lines.Count) { $lines[$i + 2] } else { "" }
            if ($line -match '(?i)subsumed\s+by' -or $next1 -match '(?i)subsumed\s+by' -or $next2 -match '(?i)subsumed\s+by') {
                $isSub = $true
            }
        } elseif ($line -match '^-\s+\*\*(FR-[A-Z]{3}-\d{3}[a-z]?)') {
            $ownedId = $Matches[1]
            if ($line -match '(?i)subsumed\s+by') { $isSub = $true }
        } elseif ($line -match '^\|\s*(?:\*\*|`)*(FR-[A-Z]{3}-\d{3}[a-z]?)(?:\*\*|`)*\s*\|') {
            $candidate = $Matches[1]
            # Only claim table-row definition if not already owned by an earlier FRD or earlier heading
            $alreadyOwnedElsewhere = $false
            foreach ($k in $frdOwnedReqs.Keys) {
                if ($k -ne $fn -and $frdOwnedReqs[$k].Contains($candidate)) {
                    $alreadyOwnedElsewhere = $true
                    break
                }
            }
            if (-not $alreadyOwnedElsewhere) {
                $ownedId = $candidate
                if ($line -match '(?i)subsumed\s+by') { $isSub = $true }
            }
        }

        if ($ownedId) {
            if (-not $frdOwnedReqs[$fn].Contains($ownedId)) {
                $frdOwnedReqs[$fn].Add($ownedId)
            }
            if ($isSub) {
                $subsumedReqs[$ownedId] = $true
            }
        }
    }
}

Write-Host "  Master FRD Catalog: $($allReqDefs.Count) defined requirements across $($existingFrds.Count) FRDs.`n" -ForegroundColor Gray

# Step 2: Validate each target TDD
foreach ($tddPath in $targetFiles) {
    $tddName = Split-Path $tddPath -Leaf
    Write-Host "----------------------------------------------------------" -ForegroundColor DarkCyan
    Write-Host "  Auditing TDD: $tddName" -ForegroundColor Cyan
    Write-Host "----------------------------------------------------------" -ForegroundColor DarkCyan

    $lines = [System.IO.File]::ReadAllLines($tddPath, [System.Text.Encoding]::UTF8)
    $fullText = [System.IO.File]::ReadAllText($tddPath, [System.Text.Encoding]::UTF8)

    # 2.1 Header Metadata Validation
    $sourceFrdRaw = $null
    $frdCommitHash = $null
    $tddVersion = $null
    $ghMilestone = $null

    foreach ($line in ($lines | Select-Object -First 40)) {
        if ($line -match '^\s*-\s+\*\*Source FRD:\*\*\s*(.+)$') { $sourceFrdRaw = $Matches[1].Trim() }
        if ($line -match '^\s*-\s+\*\*FRD Commit Hash:\*\*\s*`?([^`\s]+)`?') { $frdCommitHash = $Matches[1].Trim() }
        if ($line -match '^\s*-\s+\*\*TDD Version:\*\*\s*`?([^`\r\n]+)`?') { $tddVersion = $Matches[1].Trim() }
        if ($line -match '^\s*-\s+\*\*GitHub Milestone:\*\*\s*`?([^`\r\n]+)`?') { $ghMilestone = $Matches[1].Trim() }
    }

    if (-not $sourceFrdRaw) {
        Write-Host "  [ERROR] Missing required header bullet: '- **Source FRD:**'" -ForegroundColor Red
        $hasErrors = $true
    }
    if (-not $frdCommitHash) {
        Write-Host "  [ERROR] Missing required header bullet: '- **FRD Commit Hash:**'" -ForegroundColor Red
        $hasErrors = $true
    }
    if (-not $tddVersion) {
        Write-Host "  [ERROR] Missing required header bullet: '- **TDD Version:**'" -ForegroundColor Red
        $hasErrors = $true
    }
    if (-not $ghMilestone) {
        Write-Host "  [ERROR] Missing required header bullet: '- **GitHub Milestone:**'" -ForegroundColor Red
        $hasErrors = $true
    }

    # Resolve Source FRD filename
    $sourceFrdFileName = $null
    if ($sourceFrdRaw) {
        $frdFileMatch = [regex]::Match($sourceFrdRaw, '(\d{2}_[A-Za-z0-9_]+_FRD\.md)')
        if ($frdFileMatch.Success) {
            $sourceFrdFileName = $frdFileMatch.Groups[1].Value
            $sourceFrdFullPath = Join-Path $frdDir $sourceFrdFileName
            if (-not (Test-Path $sourceFrdFullPath)) {
                Write-Host "  [ERROR] Source FRD file does not exist in docs/frd/: $sourceFrdFileName" -ForegroundColor Red
                $hasErrors = $true
                $sourceFrdFileName = $null
            } else {
                Write-Host "  Source FRD: $sourceFrdFileName (TDD $tddVersion | Audited FRD Commit: $frdCommitHash)" -ForegroundColor Gray
                # Check FRD commit freshness if git is available
                if ($frdCommitHash -and $frdCommitHash -match '^[0-9a-f]{7,40}$') {
                    & git -C $RepoRoot cat-file -e "$frdCommitHash^{commit}" 2>$null
                    if ($LASTEXITCODE -eq 0) {
                        $newerCommits = @(& git -C $RepoRoot log --format="%h %s" "$frdCommitHash..HEAD" -- "docs/frd/$sourceFrdFileName" 2>$null)
                        if ($newerCommits.Count -gt 0) {
                            Write-Host "  [WARNING] Source FRD ($sourceFrdFileName) has $($newerCommits.Count) newer commit(s) since TDD's FRD Commit Hash ($frdCommitHash):" -ForegroundColor Yellow
                            $newerCommits | Select-Object -First 3 | ForEach-Object { Write-Host "    -> $_" -ForegroundColor Yellow }
                        }
                    }
                }
            }
        } else {
            Write-Host "  [ERROR] Could not parse valid FRD filename (XX_*_FRD.md) from Source FRD header: $sourceFrdRaw" -ForegroundColor Red
            $hasErrors = $true
        }
    }

    # 2.2 Zero Dangling FR-XXX-YYY References Check
    $tddReqRefs = @()
    for ($i = 0; $i -lt $lines.Count; $i++) {
        foreach ($m in [regex]::Matches($lines[$i], 'FR-[A-Z]{3}-\d{3}[a-z]?')) {
            $tddReqRefs += [PSCustomObject]@{ Req = $m.Value; Line = ($i + 1) }
        }
    }
    $uniqueTddRefs = @($tddReqRefs | Select-Object -ExpandProperty Req -Unique | Sort-Object)
    $danglingCount = 0
    foreach ($r in $uniqueTddRefs) {
        if (-not $allReqDefs.ContainsKey($r)) {
            $danglingCount++
            $hasErrors = $true
            $lineNums = ($tddReqRefs | Where-Object { $_.Req -eq $r } | Select-Object -ExpandProperty Line) -join ', '
            Write-Host "  [DANGLING REFERENCE] $r (lines: $lineNums) is not defined in any FRD." -ForegroundColor Red
        }
    }
    if ($danglingCount -eq 0) {
        Write-Host "  [OK] Cross-References: $($uniqueTddRefs.Count) unique FR-XXX-YYY IDs referenced, 0 dangling." -ForegroundColor Green
    }

    # 2.3 Parse Deliverable Slices (### [SLICE-XX-YY] Title)
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
            if ($slices.Contains($currentSliceId)) {
                Write-Host "  [ERROR] Duplicate Deliverable Slice heading [$currentSliceId] at line $($i + 1)." -ForegroundColor Red
                $hasErrors = $true
            } else {
                $sliceOrder.Add($currentSliceId)
                $slices[$currentSliceId] = @{
                    Id                   = $currentSliceId
                    Title                = $sliceTitle
                    Line                 = ($i + 1)
                    DeclaredId           = $null
                    Labels               = $null
                    DependsOnRaw         = $null
                    DependsOnList        = @()
                    RequirementsCovered  = @()
                    TddSections          = $null
                    TargetFiles          = New-Object System.Collections.Generic.List[string]
                    InTargetFiles        = $false
                    HasBlueprintSection  = $false
                    BlueprintLines       = 0
                    HasAcceptanceSection = $false
                    AcceptanceCheckboxes = 0
                    HasVerifySection     = $false
                    VerifyLines          = 0
                }
            }
            continue
        }

        # Exit slice parsing if a new top-level ## section starts after Section 6
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
                    $s.HasBlueprintSection = $true
                    $currentSubSection = 'blueprint'
                } elseif ($subTitle -match '(?i)^Acceptance Criteria') {
                    $s.HasAcceptanceSection = $true
                    $currentSubSection = 'acceptance'
                } elseif ($subTitle -match '(?i)^Verification Commands') {
                    $s.HasVerifySection = $true
                    $currentSubSection = 'verify'
                } else {
                    $currentSubSection = 'other'
                }
                continue
            }

            if (-not $currentSubSection) {
                if ($line -match '^\s*-\s+\*\*Slice ID:\*\*\s*`?([^`\r\n]+)`?') {
                    $s.DeclaredId = $Matches[1].Trim()
                    $s.InTargetFiles = $false
                } elseif ($line -match '^\s*-\s+\*\*Labels:\*\*\s*(.+)$') {
                    $s.Labels = $Matches[1].Trim()
                    $s.InTargetFiles = $false
                } elseif ($line -match '^\s*-\s+\*\*Depends On:\*\*\s*(.+)$') {
                    $s.DependsOnRaw = $Matches[1].Trim()
                    $s.InTargetFiles = $false
                    if ($s.DependsOnRaw -notmatch '(?i)^`?None`?$') {
                        $deps = @([regex]::Matches($s.DependsOnRaw, 'SLICE-\d{2}-\d{2}') | ForEach-Object { $_.Value })
                        $s.DependsOnList = $deps
                    }
                } elseif ($line -match '^\s*-\s+\*\*Requirements Covered:\*\*\s*(.+)$') {
                    $reqs = @([regex]::Matches($Matches[1], 'FR-[A-Z]{3}-\d{3}[a-z]?') | ForEach-Object { $_.Value })
                    $s.RequirementsCovered = $reqs
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
                if ($currentSubSection -eq 'blueprint' -and $line.Trim() -ne "") {
                    $s.BlueprintLines++
                } elseif ($currentSubSection -eq 'acceptance' -and $line -match '^\s*-\s+\[\s*[xX ]?\s*\]\s+\S+') {
                    $s.AcceptanceCheckboxes++
                } elseif ($currentSubSection -eq 'verify' -and $line.Trim() -ne "" -and $line -notmatch '^```') {
                    $s.VerifyLines++
                }
            }
        }
    }

    if ($sliceOrder.Count -eq 0) {
        Write-Host "  [ERROR] No Deliverable Slice blocks ('### [SLICE-XX-YY] Title') found in Section 6." -ForegroundColor Red
        $hasErrors = $true
    } else {
        $sliceErrors = 0
        foreach ($sid in $sliceOrder) {
            $s = $slices[$sid]
            if ($s.DeclaredId -ne $sid) {
                Write-Host "  [ERROR] Slice [$sid] (line $($s.Line)): '- **Slice ID:**' is '$($s.DeclaredId)', expected '$sid'." -ForegroundColor Red
                $sliceErrors++; $hasErrors = $true
            }
            if (-not $s.Labels) {
                Write-Host "  [ERROR] Slice [$sid] (line $($s.Line)): Missing '- **Labels:**'." -ForegroundColor Red
                $sliceErrors++; $hasErrors = $true
            }
            if (-not $s.DependsOnRaw) {
                Write-Host "  [ERROR] Slice [$sid] (line $($s.Line)): Missing '- **Depends On:**' (use 'None' or upstream SLICE-XX-YY IDs)." -ForegroundColor Red
                $sliceErrors++; $hasErrors = $true
            } elseif ($s.DependsOnRaw -notmatch '(?i)^`?None`?$' -and $s.DependsOnList.Count -eq 0) {
                Write-Host "  [ERROR] Slice [$sid] (line $($s.Line)): '- **Depends On:**' must be 'None' or valid SLICE-XX-YY ID(s), got '$($s.DependsOnRaw)'." -ForegroundColor Red
                $sliceErrors++; $hasErrors = $true
            }
            if ($s.RequirementsCovered.Count -eq 0) {
                Write-Host "  [ERROR] Slice [$sid] (line $($s.Line)): '- **Requirements Covered:**' must cite at least one FR-XXX-YYY." -ForegroundColor Red
                $sliceErrors++; $hasErrors = $true
            }
            if (-not $s.TddSections) {
                Write-Host "  [ERROR] Slice [$sid] (line $($s.Line)): Missing '- **TDD Sections:**'." -ForegroundColor Red
                $sliceErrors++; $hasErrors = $true
            }
            if ($s.TargetFiles.Count -eq 0) {
                Write-Host "  [ERROR] Slice [$sid] (line $($s.Line)): '- **Target Files:**' must list at least one target file bullet." -ForegroundColor Red
                $sliceErrors++; $hasErrors = $true
            }
            if (-not $s.HasBlueprintSection -or $s.BlueprintLines -lt 2) {
                Write-Host "  [ERROR] Slice [$sid] (line $($s.Line)): Missing or empty '#### Implementation Blueprint & Signatures' section." -ForegroundColor Red
                $sliceErrors++; $hasErrors = $true
            }
            if (-not $s.HasAcceptanceSection -or $s.AcceptanceCheckboxes -eq 0) {
                Write-Host "  [ERROR] Slice [$sid] (line $($s.Line)): '#### Acceptance Criteria' must include at least one '- [ ]' checkbox." -ForegroundColor Red
                $sliceErrors++; $hasErrors = $true
            }
            if (-not $s.HasVerifySection -or $s.VerifyLines -eq 0) {
                Write-Host "  [ERROR] Slice [$sid] (line $($s.Line)): Missing or empty '#### Verification Commands' section." -ForegroundColor Red
                $sliceErrors++; $hasErrors = $true
            }

            # Check Depends On targets exist and no self-dependency
            foreach ($dep in $s.DependsOnList) {
                if ($dep -eq $sid) {
                    Write-Host "  [ERROR] Slice [$sid] depends on itself." -ForegroundColor Red
                    $sliceErrors++; $hasErrors = $true
                } elseif (-not $slices.Contains($dep)) {
                    Write-Host "  [ERROR] Slice [$sid] depends on unknown slice '$dep'." -ForegroundColor Red
                    $sliceErrors++; $hasErrors = $true
                }
            }
        }

        # Check DAG for cycles via DFS (0=unvisited, 1=visiting, 2=visited)
        $visitState = @{}
        foreach ($sid in $sliceOrder) { $visitState[$sid] = 0 }
        $hasCycle = $false

        function Test-SliceCycle([string]$NodeId) {
            $visitState[$NodeId] = 1
            foreach ($dep in $slices[$NodeId].DependsOnList) {
                if (-not $slices.Contains($dep)) { continue }
                if ($visitState[$dep] -eq 1) {
                    Write-Host "  [ERROR] Dependency cycle detected involving [$NodeId] -> [$dep]." -ForegroundColor Red
                    $script:hasCycle = $true
                    $script:hasErrors = $true
                } elseif ($visitState[$dep] -eq 0) {
                    Test-SliceCycle $dep
                }
            }
            $visitState[$NodeId] = 2
        }

        foreach ($sid in $sliceOrder) {
            if ($visitState[$sid] -eq 0) {
                Test-SliceCycle $sid
            }
        }

        if ($sliceErrors -eq 0 -and -not $hasCycle) {
            Write-Host "  [OK] Deliverable Slice Manifest: $($sliceOrder.Count) valid slices with acyclic dependency DAG." -ForegroundColor Green
        }
    }

    # 2.4 Validate 100% Source FRD Requirement Coverage (Traceability Matrix + Slices)
    if ($sourceFrdFileName -and $frdOwnedReqs.ContainsKey($sourceFrdFileName)) {
        $ownedList = $frdOwnedReqs[$sourceFrdFileName]
        $activeReqs = @($ownedList | Where-Object { -not $subsumedReqs.ContainsKey($_) })
        $subsumedList = @($ownedList | Where-Object { $subsumedReqs.ContainsKey($_) })

        # Collect requirements covered across all slices
        $sliceCoveredSet = @{}
        foreach ($sid in $sliceOrder) {
            foreach ($r in $slices[$sid].RequirementsCovered) {
                $sliceCoveredSet[$r] = $true
            }
        }

        # Collect requirements present in markdown table rows (Traceability Matrix)
        $matrixCoveredSet = @{}
        foreach ($line in $lines) {
            if ($line -match '^\|') {
                foreach ($m in [regex]::Matches($line, 'FR-[A-Z]{3}-\d{3}[a-z]?')) {
                    $matrixCoveredSet[$m.Value] = $true
                }
            }
        }

        $missingInMatrix = @()
        $missingInSlices = @()
        $missingSubsumed = @()

        foreach ($req in $activeReqs) {
            if (-not $matrixCoveredSet.ContainsKey($req)) {
                $missingInMatrix += $req
            }
            if (-not $sliceCoveredSet.ContainsKey($req)) {
                $missingInSlices += $req
            }
        }

        foreach ($subReq in $subsumedList) {
            if (-not $uniqueTddRefs.Contains($subReq)) {
                $missingSubsumed += $subReq
            }
        }

        if ($missingInMatrix.Count -gt 0) {
            Write-Host "  [ERROR] Active FRD requirements missing from TDD Traceability Matrix ($($missingInMatrix.Count)):" -ForegroundColor Red
            Write-Host "    -> $($missingInMatrix -join ', ')" -ForegroundColor DarkRed
            $hasErrors = $true
        }
        if ($missingInSlices.Count -gt 0) {
            Write-Host "  [ERROR] Active FRD requirements not assigned to any Deliverable Slice ($($missingInSlices.Count)):" -ForegroundColor Red
            Write-Host "    -> $($missingInSlices -join ', ')" -ForegroundColor DarkRed
            $hasErrors = $true
        }
        if ($missingSubsumed.Count -gt 0) {
            Write-Host "  [ERROR] Subsumed FRD requirements not accounted for in TDD Traceability Matrix ($($missingSubsumed.Count)):" -ForegroundColor Red
            Write-Host "    -> $($missingSubsumed -join ', ')" -ForegroundColor DarkRed
            $hasErrors = $true
        }

        if ($missingInMatrix.Count -eq 0 -and $missingInSlices.Count -eq 0 -and $missingSubsumed.Count -eq 0) {
            Write-Host "  [OK] 100% Source FRD Coverage: $($activeReqs.Count)/$($activeReqs.Count) Active requirements sliced & traced; $($subsumedList.Count)/$($subsumedList.Count) Subsumed accounted for." -ForegroundColor Green
        }
    }
    Write-Host ""
}

if (-not $hasErrors) {
    Write-Host "  [PASS] All TDD coverage, cross-reference, and slice DAG checks passed!" -ForegroundColor Green
    Write-Host "==========================================================" -ForegroundColor Cyan
    exit 0
} else {
    Write-Host "  [FAIL] TDD linting failed. Resolve the errors above before committing or publishing issues." -ForegroundColor Red
    Write-Host "==========================================================" -ForegroundColor Cyan
    exit 1
}
