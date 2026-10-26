<#
.SYNOPSIS
Builds one wave unattended (lead-architect Runbook D7).

.DESCRIPTION
Opens one fresh headless lead-architect session per task. Each session moves
one task and merges it into an autoloop/ run branch, never the integration
branch. After each session the lead adds one row to the run file, and its last
cell decides what happens next: CONTINUE, WAVE-COMPLETE or STOP: <reason>.
The runner also stops on its own budgets (tasks, spend per session, hours),
on a failed session, on any refused tool call, and on a missing or extra run
row. It never merges the run branch: the user does.

.EXAMPLE
go tool task wave-run -- -Wave B -Tasks 3
go tool task wave-run -- -Wave B -DryRun
#>
param(
    [Parameter(Mandatory = $true)][string]$Wave,
    [int]$Tasks = 3,
    [double]$MaxUsdPerTask = 40,
    [double]$MaxHours = 8,
    [string]$Milestone = 'm2',
    [string]$Base = 'main',
    [string]$Effort = '',
    [string]$Claude = '',
    [switch]$DryRun
)

# Not 'Stop': in Windows PowerShell a native command's stderr then becomes a
# terminating error. Native exit codes are checked by hand instead.
$ErrorActionPreference = 'Continue'
Set-Location (git rev-parse --show-toplevel).Trim()

$branch = "autoloop/$Milestone-wave-$($Wave.ToLower())"
$slug = $branch -replace '/', '-'
$runFile = ".agents/tasks/$Milestone/runs/$slug.md"
$logDir = Join-Path $env:TEMP "wave-run\$slug"
$started = Get-Date
$spent = 0.0

function Find-Claude {
    if ($Claude) { return $Claude }
    $cmd = Get-Command claude -ErrorAction SilentlyContinue
    if ($cmd) { return $cmd.Source }
    # The IDE extensions ship the CLI without putting it on PATH.
    Get-ChildItem -Path "$env:USERPROFILE\.antigravity-ide\extensions", "$env:USERPROFILE\.vscode\extensions" `
        -Filter 'anthropic.claude-code-*' -Directory -ErrorAction SilentlyContinue |
        Sort-Object LastWriteTime -Descending |
        ForEach-Object { Join-Path $_.FullName 'resources\native-binary\claude.exe' } |
        Where-Object { Test-Path $_ } |
        Select-Object -First 1
}

function Test-BranchExists([string]$name) {
    git rev-parse --verify --quiet "refs/heads/$name" | Out-Null
    return $LASTEXITCODE -eq 0
}

# The run file as committed on the run branch, or nothing before the first row.
function Get-RunText {
    if (-not (Test-BranchExists $branch)) { return @() }
    if (-not (git ls-tree --name-only $branch -- $runFile)) { return @() }
    @(git show "${branch}:$runFile")
}

function Get-RunRows { @(Get-RunText | Where-Object { $_ -match '^\|\s*\d+\s*\|' }) }

function Get-Signal([string]$row) {
    $cells = $row.Trim().Trim('|').Split('|')
    $cells[$cells.Count - 1].Trim()
}

# Wave cards from the run file header still listed in NotYetExecutable.
function Get-PendingCards {
    $line = Get-RunText | Where-Object { $_ -match '^- \*\*Cards:\*\*' } | Select-Object -First 1
    if (-not $line) { return @('(run file has no Cards line)') }
    $names = ($line -replace '^- \*\*Cards:\*\*\s*', '').Split(';') | ForEach-Object { $_.Trim() } | Where-Object { $_ }
    $listed = @(git show "${branch}:pkg/engine/cards/cards.go") | ForEach-Object { $_.Trim() }
    @($names | Where-Object { $listed -contains "`"$_`"," })
}

function Test-Gates {
    $wt = Join-Path $env:TEMP "wave-run-verify-$PID"
    git worktree add --detach $wt $branch | Out-Null
    if ($LASTEXITCODE -ne 0) { return $false }
    go -C $wt tool task check
    $ok = $LASTEXITCODE -eq 0
    if ($ok) {
        go -C $wt tool task sim
        $ok = $LASTEXITCODE -eq 0
    }
    git worktree remove --force $wt | Out-Null
    return $ok
}

function Stop-Run([string]$reason, [int]$code = 1) {
    $hours = [math]::Round(((Get-Date) - $started).TotalHours, 2)
    Write-Host ''
    Write-Host "== Wave run stopped: $reason"
    Write-Host "   Spent: `$$([math]::Round($spent, 2)) over $hours h. Session logs: $logDir"
    Write-Host "   Review:  git log --oneline $Base..$branch"
    Write-Host "            git show ${branch}:$runFile"
    Write-Host "   Land it: git merge --no-ff $branch   (on $Base, when you are satisfied)"
    Write-Host "   Resume:  go tool task wave-run -- -Wave $Wave -Tasks <n>"
    exit $code
}

$exe = Find-Claude
if (-not $exe) { Write-Host 'claude CLI not found; pass -Claude <path to claude.exe>'; exit 1 }
if (-not (Test-BranchExists $Base)) { Write-Host "Base branch $Base does not exist"; exit 1 }

$cliArgs = @(
    '--model', 'opus',
    '--permission-mode', 'auto',
    '--permission-prompts', 'none',
    '--max-budget-usd', $MaxUsdPerTask.ToString([Globalization.CultureInfo]::InvariantCulture),
    '--output-format', 'json'
)
if ($Effort) { $cliArgs += @('--effort', $Effort) }
$cliArgs += @('--disallowedTools', 'Bash(git push *),PowerShell(git push *)')

if ($DryRun) {
    Write-Host "Run branch: $branch (exists: $(Test-BranchExists $branch))"
    Write-Host "Run file:   $runFile ($((Get-RunRows).Count) rows)"
    Write-Host "Claude:     $exe"
    Write-Host "Session:    claude -p `"@lead-architect next (wave run: ...)`" $($cliArgs -join ' ')"
    exit 0
}

if (-not (Test-BranchExists $branch)) {
    git branch $branch $Base
    if ($LASTEXITCODE -ne 0) { Write-Host "Could not create $branch"; exit 1 }
    Write-Host "Created $branch from $Base."
}
New-Item -ItemType Directory -Force $logDir | Out-Null

for ($n = 1; $n -le $Tasks; $n++) {
    if (((Get-Date) - $started).TotalHours -ge $MaxHours) { Stop-Run "time budget of $MaxHours h used up" }

    $before = (Get-RunRows).Count
    $prompt = "@lead-architect next (wave run: wave $Wave, run branch $branch, run file $runFile)"
    $log = Join-Path $logDir ('{0:yyyyMMdd-HHmmss}-{1}.json' -f (Get-Date), $n)
    Write-Host "[$n/$Tasks] $(Get-Date -Format HH:mm) lead session on $branch (log: $log)"

    $out = & $exe -p $prompt -n "wave-run $Wave #$n" @cliArgs
    $exit = $LASTEXITCODE
    $out | Set-Content -Encoding utf8 $log

    try { $res = ($out -join "`n") | ConvertFrom-Json } catch { $res = $null }
    if ($res -and $res.total_cost_usd) { $spent += [double]$res.total_cost_usd }
    if ($exit -ne 0 -or -not $res) { Stop-Run "session $n exited with code $exit" }
    if ($res.is_error) { Stop-Run "session $n ended with $($res.subtype)" }
    if (@($res.permission_denials).Count -gt 0) {
        $res.permission_denials | ForEach-Object { Write-Host "   refused: $($_.tool_name) $($_.tool_input | ConvertTo-Json -Compress)" }
        Stop-Run "session $n had refused tool calls"
    }

    $rows = Get-RunRows
    if ($rows.Count -ne $before + 1) { Stop-Run "session $n added $($rows.Count - $before) run rows, expected 1" }
    $row = $rows[$rows.Count - 1]
    Write-Host "   $row"

    $signal = Get-Signal $row
    if ($signal -eq 'CONTINUE') { continue }
    if ($signal -eq 'WAVE-COMPLETE') {
        $pending = Get-PendingCards
        if ($pending.Count -gt 0) { Stop-Run "lead says complete, but still not executable: $($pending -join '; ')" }
        Write-Host '   Verifying gates on the run branch...'
        if (-not (Test-Gates)) { Stop-Run 'lead says complete, but check or sim fails on the run branch' }
        Stop-Run "wave $Wave complete" 0
    }
    if ($signal -like 'STOP*') { Stop-Run "lead: $signal" }
    Stop-Run "unknown signal '$signal'"
}
Stop-Run "task budget of $Tasks used up" 0
