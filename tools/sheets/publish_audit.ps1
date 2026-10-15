<#
.SYNOPSIS
    Publishes the Card Town Card Standardization Audit to Google Drive as both an
    interactive, formatted Google Sheet and a native Google Doc, plus synced local
    Markdown and HTML reports.

.DESCRIPTION
    Reads tools/linter/card_standardization_audit.json and publishes:
    1. A native Google Sheet "Card Standardization Audit" with 8 structured tabs:
       - Summary Dashboard (KPIs, faction breakdown, Domain 7.1 high priority review items)
       - Flagged for Review (all 52 cards flagged for designer sign-off, with high-priority items first)
       - Master Audit Log (all 167 audited cards with full diffs and notes)
       - Animus (42 cards)
       - Concordia (41 cards)
       - Path (42 cards)
       - Soulless (41 cards)
       - HQ & Neutral (9 cards)
    2. A native Google Doc "Card Standardization Audit Report" formatted for review and commenting.
    3. Synced local files written directly to the Google Drive folder:
       - Card_Standardization_Audit_Report.md
       - Card_Standardization_Audit_Report.html
    4. Local repository copy in docs/Card_Standardization_Audit_Report.md.

.EXAMPLE
    powershell -ExecutionPolicy Bypass -File tools\sheets\publish_audit.ps1

.EXAMPLE
    # Preview without contacting Google
    powershell -ExecutionPolicy Bypass -File tools\sheets\publish_audit.ps1 -DryRun
#>
param(
    [string]$RepoRoot = (Resolve-Path "$PSScriptRoot\..\..").Path,
    [string]$DestinationPath = "G:\My Drive\The End of the Manufacture\Card Town\Card sheets",
    [string]$ConfigDir = (Join-Path $env:APPDATA "animus-mundi"),
    [string]$ClientSecretPath,
    [switch]$DryRun,
    [switch]$ReAuth
)

$ErrorActionPreference = 'Stop'
[Net.ServicePointManager]::SecurityProtocol = [Net.ServicePointManager]::SecurityProtocol -bor [Net.SecurityProtocolType]::Tls12

$auditJsonPath = Join-Path $RepoRoot "tools\linter\card_standardization_audit.json"
if (-not (Test-Path $auditJsonPath)) {
    throw "Audit file not found at '$auditJsonPath'. Run tools\linter\standardize_all_cards.ps1 first."
}

if (-not $ClientSecretPath) { $ClientSecretPath = Join-Path $ConfigDir "client_secret.json" }
$TokenPath = Join-Path $ConfigDir "google-refresh-token.dat"
$Scopes = "https://www.googleapis.com/auth/drive"

$highPriorityNames = @(
    "Animoceps Rager",
    "Funeral Fungus",
    "Rift",
    "Mutation Swell",
    "Steel Rain",
    "Macrocannon Barrage",
    "Reverberating Terror",
    "Devilhorn Guerrillas",
    "All-Out War"
)

# -----------------------------------------------------------------------------
# Google OAuth Helpers
# -----------------------------------------------------------------------------
function Get-ClientConfig {
    if (-not (Test-Path $ClientSecretPath) -and (Test-Path "$ClientSecretPath.json")) {
        $script:ClientSecretPath = "$ClientSecretPath.json"
    }
    if (-not (Test-Path $ClientSecretPath)) {
        throw "Google OAuth client file not found at '$ClientSecretPath'. See tools\sheets\README.md for setup instructions."
    }
    $json = Get-Content $ClientSecretPath -Raw | ConvertFrom-Json
    if (-not $json.installed) {
        throw "'$ClientSecretPath' is not a 'Desktop app' OAuth client."
    }
    return $json.installed
}

function ConvertTo-Base64Url([byte[]]$bytes) {
    return [Convert]::ToBase64String($bytes).TrimEnd('=').Replace('+', '-').Replace('/', '_')
}

function Save-RefreshToken([string]$token) {
    New-Item -ItemType Directory -Path $ConfigDir -Force | Out-Null
    ConvertTo-SecureString $token -AsPlainText -Force | ConvertFrom-SecureString | Set-Content -Path $TokenPath
}

function Read-RefreshToken {
    $secure = (Get-Content $TokenPath -Raw).Trim() | ConvertTo-SecureString
    $bstr = [Runtime.InteropServices.Marshal]::SecureStringToBSTR($secure)
    try { return [Runtime.InteropServices.Marshal]::PtrToStringBSTR($bstr) }
    finally { [Runtime.InteropServices.Marshal]::ZeroFreeBSTR($bstr) }
}

function Request-NewTokens($client) {
    $rng = [Security.Cryptography.RandomNumberGenerator]::Create()
    $verifierBytes = New-Object byte[] 48
    $rng.GetBytes($verifierBytes)
    $verifier = ConvertTo-Base64Url $verifierBytes
    $challenge = ConvertTo-Base64Url ([Security.Cryptography.SHA256]::Create().ComputeHash([Text.Encoding]::ASCII.GetBytes($verifier)))
    $state = [guid]::NewGuid().ToString("N")

    $probe = New-Object Net.Sockets.TcpListener([Net.IPAddress]::Loopback, 0)
    $probe.Start(); $port = $probe.LocalEndpoint.Port; $probe.Stop()
    $redirect = "http://localhost:$port/"

    $listener = New-Object Net.HttpListener
    $listener.Prefixes.Add($redirect)
    $listener.Start()
    try {
        $query = @(
            "client_id=$([Uri]::EscapeDataString($client.client_id))",
            "redirect_uri=$([Uri]::EscapeDataString($redirect))",
            "response_type=code",
            "scope=$([Uri]::EscapeDataString($Scopes))",
            "code_challenge=$challenge",
            "code_challenge_method=S256",
            "access_type=offline",
            "prompt=consent",
            "state=$state"
        ) -join "&"
        $authUrl = "https://accounts.google.com/o/oauth2/v2/auth?$query"
        Write-Host "Opening your browser for Google sign-in. If it does not open, visit:`n$authUrl`n" -ForegroundColor Cyan
        Start-Process $authUrl

        $deadline = (Get-Date).AddMinutes(5)
        $code = $null
        while (-not $code) {
            $async = $listener.BeginGetContext($null, $null)
            while (-not $async.AsyncWaitHandle.WaitOne(500)) {
                if ((Get-Date) -gt $deadline) { throw "Timed out waiting for Google sign-in." }
            }
            $ctx = $listener.EndGetContext($async)
            $q = $ctx.Request.QueryString
            if (-not $q["code"] -and -not $q["error"]) {
                $ctx.Response.StatusCode = 404; $ctx.Response.Close(); continue
            }
            $msg = if ($q["code"]) { "Signed in. You can close this tab and return to the terminal." } else { "Sign-in failed: $($q['error'])" }
            $buf = [Text.Encoding]::UTF8.GetBytes("<html><body style='font-family:sans-serif'><h3>$msg</h3></body></html>")
            $ctx.Response.ContentType = "text/html; charset=utf-8"
            $ctx.Response.OutputStream.Write($buf, 0, $buf.Length)
            $ctx.Response.Close()
            if ($q["error"]) { throw "Google sign-in failed: $($q['error'])" }
            if ($q["state"] -ne $state) { throw "Google sign-in failed: OAuth state mismatch." }
            $code = $q["code"]
        }
    } finally {
        $listener.Stop(); $listener.Close()
    }

    return Invoke-RestMethod -Method Post -Uri "https://oauth2.googleapis.com/token" -Body @{
        client_id     = $client.client_id
        client_secret = $client.client_secret
        code          = $code
        code_verifier = $verifier
        grant_type    = "authorization_code"
        redirect_uri  = $redirect
    }
}

function Get-AccessToken {
    $client = Get-ClientConfig
    if (-not $ReAuth -and (Test-Path $TokenPath)) {
        try {
            $resp = Invoke-RestMethod -Method Post -Uri "https://oauth2.googleapis.com/token" -Body @{
                client_id     = $client.client_id
                client_secret = $client.client_secret
                refresh_token = (Read-RefreshToken)
                grant_type    = "refresh_token"
            }
            return $resp.access_token
        } catch {
            Write-Warning "Saved Google sign-in is no longer valid; signing in again."
        }
    }
    $resp = Request-NewTokens $client
    if ($resp.refresh_token) { Save-RefreshToken $resp.refresh_token }
    return $resp.access_token
}

# -----------------------------------------------------------------------------
# Google API Helpers
# -----------------------------------------------------------------------------
function Invoke-GoogleApi([string]$Method, [string]$Uri, $Body = $null) {
    $params = @{ Method = $Method; Uri = $Uri; Headers = @{ Authorization = "Bearer $script:AccessToken" } }
    if ($null -ne $Body) {
        $json = ConvertTo-Json -InputObject $Body -Depth 50 -Compress
        $params.Body = [Text.Encoding]::UTF8.GetBytes($json)
        $params.ContentType = "application/json; charset=utf-8"
    }
    try { return Invoke-RestMethod @params }
    catch { throw "Google API $Method $Uri failed: $($_.Exception.Message) $($_.ErrorDetails.Message)" }
}

function Format-DriveQueryString([string]$s) { return $s.Replace('\', '\\').Replace("'", "\'") }

function Find-DriveFiles([string]$query) {
    $uri = "https://www.googleapis.com/drive/v3/files?q=" + [Uri]::EscapeDataString($query) + "&fields=files(id,name,mimeType)&spaces=drive"
    return @((Invoke-GoogleApi GET $uri).files)
}

function Resolve-DriveFolderId([string]$localPath) {
    $marker = "\My Drive\"
    $idx = $localPath.IndexOf($marker, [StringComparison]::OrdinalIgnoreCase)
    if ($idx -lt 0) { throw "DestinationPath '$localPath' must be inside a Google Drive for Desktop 'My Drive' folder." }
    $parentId = "root"
    $walked = "My Drive"
    foreach ($segment in ($localPath.Substring($idx + $marker.Length) -split '\\' | Where-Object { $_ })) {
        $found = Find-DriveFiles "'$parentId' in parents and name = '$(Format-DriveQueryString $segment)' and mimeType = 'application/vnd.google-apps.folder' and trashed = false"
        if ($found.Count -eq 0) { throw "Drive folder '$segment' not found in '$walked'." }
        if ($found.Count -gt 1) { Write-Warning "Multiple Drive folders named '$segment' in '$walked'; using the first." }
        $parentId = $found[0].id
        $walked = "$walked\$segment"
    }
    return $parentId
}

function Get-OrCreateSpreadsheet([string]$folderId, [string]$name) {
    $found = Find-DriveFiles "'$folderId' in parents and name = '$(Format-DriveQueryString $name)' and mimeType = 'application/vnd.google-apps.spreadsheet' and trashed = false"
    if ($found.Count -ge 1) { return @{ Id = $found[0].id; Created = $false } }

    $file = Invoke-GoogleApi POST "https://www.googleapis.com/drive/v3/files?fields=id" @{
        name     = $name
        mimeType = "application/vnd.google-apps.spreadsheet"
        parents  = @($folderId)
    }
    return @{ Id = $file.id; Created = $true }
}

function Format-A1Sheet([string]$title) { return "'" + $title.Replace("'", "''") + "'" }

function Get-SheetProperties([string]$base) {
    $map = @{}
    foreach ($s in @((Invoke-GoogleApi GET "$base`?fields=sheets.properties").sheets)) { $map[$s.properties.title] = $s.properties }
    return $map
}

# -----------------------------------------------------------------------------
# Data Tab Preparation
# -----------------------------------------------------------------------------
$audit = Get-Content $auditJsonPath -Raw | ConvertFrom-Json
$totalCards = $audit.Count
$changedCards = ($audit | Where-Object { $_.TextChanged }).Count
$flaggedCards = ($audit | Where-Object { $_.Flag -ne "None." }).Count
$unchangedCards = $totalCards - $changedCards

# 1. Summary Dashboard tab
$summaryRows = New-Object System.Collections.Generic.List[object]
$summaryRows.Add(@("CARD TOWN - CARD STANDARDIZATION AUDIT REPORT", "", "", "", ""))
$summaryRows.Add(@("Audit Date", "October 2026", "Source", "tools/linter/card_standardization_audit.json", ""))
$summaryRows.Add(@("Companion FRD", "Card Data Pipeline FRD (Domain 7)", "Target Enums", "FR-DAT-006 / FR-DAT-007", ""))
$summaryRows.Add(@("", "", "", "", ""))
$summaryRows.Add(@("KEY AUDIT METRICS", "COUNT", "PERCENTAGE", "DESCRIPTION", ""))
$summaryRows.Add(@("Total Cards Audited", $totalCards, "100.0%", "All cards across all 10 balance CSV sheets", ""))
$summaryRows.Add(@("Rules Text Standardized", $changedCards, ("{0:P1}" -f ($changedCards / $totalCards)), "MTG capitalization, punctuation, zones, keywords", ""))
$summaryRows.Add(@("Flagged for Designer Review", $flaggedCards, ("{0:P1}" -f ($flaggedCards / $totalCards)), "Design ambiguities and rule mechanics requiring sign-off", ""))
$summaryRows.Add(@("Rules Text Unchanged", $unchangedCards, ("{0:P1}" -f ($unchangedCards / $totalCards)), "Cards already conforming to grammar guidelines", ""))
$summaryRows.Add(@("", "", "", "", ""))
$summaryRows.Add(@("FACTION BREAKDOWN", "TOTAL CARDS", "MANEUVERS", "UNITS & STRUCTURES", "HEADQUARTERS"))
foreach ($f in @("ANIMUS", "CONCORDIA", "PATH", "SOULLESS")) {
    $fCards = $audit | Where-Object { $_.Faction -eq $f }
    $man = ($fCards | Where-Object { $_.CardType -eq "MANEUVER" }).Count
    $us = ($fCards | Where-Object { $_.CardType -in @("UNIT", "STRUCTURE") }).Count
    $hq = ($fCards | Where-Object { $_.CardType -eq "HEADQUARTERS" }).Count
    $summaryRows.Add(@($f, $fCards.Count, $man, $us, $hq))
}
$cCards = $audit | Where-Object { $_.Faction -in @("COMMON", "NEUTRAL") }
$summaryRows.Add(@("COMMON / NEUTRAL", $cCards.Count, 0, $cCards.Count, 0))
$summaryRows.Add(@("", "", "", "", ""))
$summaryRows.Add(@("HIGH-PRIORITY DESIGNER REVIEW ITEMS (DOMAIN 7.1)", "", "", "", ""))
$summaryRows.Add(@("Card Title", "Faction", "Card Type", "Design Decision / Flag", "Companion Requirement"))
$summaryRows.Add(@("Animoceps Rager", "ANIMUS", "HEADQUARTERS", "Formalized per-turn limit syntax: 'This ability triggers only once each turn.'", "FR-ABL-008"))
$summaryRows.Add(@("Funeral Fungus", "ANIMUS", "MANEUVER", "Templated exhaust as additional enactment cost: 'As an additional cost to enact this maneuver, exhaust a readied allied unit.'", "FR-ABL-009"))
$summaryRows.Add(@("Rift", "ANIMUS", "HEADQUARTERS", "Self-damage to Headquarters confirmed; Headquarters is eligible target for damage mitigation (Unbowed).", "FR-REP-004"))
$summaryRows.Add(@("Mutation Swell", "ANIMUS", "MANEUVER", "Clarified causality clause ('killed this way'); confirm odd-number rounding (floor vs error).", "FR-ABL-007"))
$summaryRows.Add(@("Steel Rain", "PATH", "MANEUVER", "Clarified damage allocation: single chosen target vs multi-target distributed pool.", "FR-ABL-011"))
$summaryRows.Add(@("Macrocannon Barrage", "SOULLESS", "MANEUVER", "Standardized distributed pool damage syntax: 'Distribute 5 damage among up to 5 chosen units.'", "FR-ABL-011"))
$summaryRows.Add(@("Reverberating Terror", "CONCORDIA", "MANEUVER", "Title spelling normalized from 'reverbating terror' to 'Reverberating Terror'.", "FR-TXT-001"))
$summaryRows.Add(@("Devilhorn Guerrillas", "PATH", "UNIT", "Title spelling normalized from 'Devilhorn Guerillas' to 'Devilhorn Guerrillas'.", "FR-TXT-001"))
$summaryRows.Add(@("All-Out War", "SOULLESS", "MANEUVER", "Title hyphenated ('All-Out War'); modal list templating standardized.", "FR-ABL-010"))

# 2. Flagged for Review tab
$flaggedRows = New-Object System.Collections.Generic.List[object]
$flaggedRows.Add(@("Priority", "Faction", "Card Title", "Card Type", "Rarity", "Designer Review Flag", "Standardization Note", "Standardized Rules Text", "Previous Rules Text", "Source File", "Designer Sign-off"))

$highCards = $audit | Where-Object { $_.StandardizedName -in $highPriorityNames } | Sort-Object Faction, CardType, StandardizedName
$standardCards = $audit | Where-Object { $_.Flag -ne "None." -and $_.StandardizedName -notin $highPriorityNames } | Sort-Object Faction, CardType, StandardizedName

foreach ($c in $highCards) {
    $flaggedRows.Add(@("HIGH PRIORITY", $c.Faction, $c.StandardizedName, $c.CardType, $c.Rarity, $c.Flag, $c.Note, $c.StandardizedText, $c.OriginalText, $c.File, ""))
}
foreach ($c in $standardCards) {
    $flaggedRows.Add(@("STANDARD REVIEW", $c.Faction, $c.StandardizedName, $c.CardType, $c.Rarity, $c.Flag, $c.Note, $c.StandardizedText, $c.OriginalText, $c.File, ""))
}

# 3. Master Audit Log tab
$masterRows = New-Object System.Collections.Generic.List[object]
$masterRows.Add(@("Faction", "Card Type", "Rarity", "Card Title", "Original Title", "Status", "Standardized Rules Text", "Previous Rules Text", "Standardization Note", "Designer Review Flag", "Source File"))

$sortedAll = $audit | Sort-Object Faction, CardType, StandardizedName
foreach ($c in $sortedAll) {
    $status = if ($c.Flag -ne "None.") { "FLAGGED FOR REVIEW" } elseif ($c.TextChanged) { "STANDARDIZED" } else { "UNCHANGED" }
    $masterRows.Add(@($c.Faction, $c.CardType, $c.Rarity, $c.StandardizedName, $c.OriginalName, $status, $c.StandardizedText, $c.OriginalText, $c.Note, $c.Flag, $c.File))
}

# 4-8. Faction Tabs
function Build-FactionTabRows($factionName) {
    $rows = New-Object System.Collections.Generic.List[object]
    $rows.Add(@("Card Type", "Rarity", "Card Title", "Original Title", "Status", "Standardized Rules Text", "Previous Rules Text", "Standardization Note", "Designer Review Flag"))
    $filtered = if ($factionName -eq "HQ & Neutral") {
        $audit | Where-Object { $_.CardType -eq "HEADQUARTERS" -or $_.Faction -in @("COMMON", "NEUTRAL") } | Sort-Object CardType, Faction, StandardizedName
    } else {
        $audit | Where-Object { $_.Faction -eq $factionName -and $_.CardType -ne "HEADQUARTERS" } | Sort-Object CardType, StandardizedName
    }
    foreach ($c in $filtered) {
        $status = if ($c.Flag -ne "None.") { "FLAGGED FOR REVIEW" } elseif ($c.TextChanged) { "STANDARDIZED" } else { "UNCHANGED" }
        $rows.Add(@($c.CardType, $c.Rarity, $c.StandardizedName, $c.OriginalName, $status, $c.StandardizedText, $c.OriginalText, $c.Note, $c.Flag))
    }
    return $rows
}

$allTabs = @(
    @{ Title = "Summary Dashboard"; Rows = $summaryRows; Cols = 5; Type = "dashboard" },
    @{ Title = "Flagged for Review"; Rows = $flaggedRows; Cols = 11; Type = "flagged" },
    @{ Title = "Master Audit Log"; Rows = $masterRows; Cols = 11; Type = "master" },
    @{ Title = "Animus"; Rows = (Build-FactionTabRows "ANIMUS"); Cols = 9; Type = "faction" },
    @{ Title = "Concordia"; Rows = (Build-FactionTabRows "CONCORDIA"); Cols = 9; Type = "faction" },
    @{ Title = "Path"; Rows = (Build-FactionTabRows "PATH"); Cols = 9; Type = "faction" },
    @{ Title = "Soulless"; Rows = (Build-FactionTabRows "SOULLESS"); Cols = 9; Type = "faction" },
    @{ Title = "HQ & Neutral"; Rows = (Build-FactionTabRows "HQ & Neutral"); Cols = 9; Type = "faction" }
)

# -----------------------------------------------------------------------------
# Google Sheets Publisher
# -----------------------------------------------------------------------------
function Publish-AuditSpreadsheet([string]$spreadsheetId, [bool]$isNew, $tabs) {
    $base = "https://sheets.googleapis.com/v4/spreadsheets/$spreadsheetId"
    $existing = Get-SheetProperties $base
    $titles = @($tabs | ForEach-Object { $_.Title })

    # 1. Structure
    $requests = @()
    foreach ($t in $tabs) {
        $needRows = [Math]::Max($t.Rows.Count, 1)
        if (-not $existing.ContainsKey($t.Title)) {
            $requests += @{ addSheet = @{ properties = @{ title = $t.Title; gridProperties = @{
                rowCount = [Math]::Max($needRows + 10, 200); columnCount = [Math]::Max($t.Cols + 2, 15) } } } }
        } else {
            $grid = $existing[$t.Title].gridProperties
            if ($grid.rowCount -lt $needRows -or $grid.columnCount -lt $t.Cols) {
                $requests += @{ updateSheetProperties = @{
                    properties = @{ sheetId = $existing[$t.Title].sheetId; gridProperties = @{
                        rowCount = [Math]::Max($grid.rowCount, $needRows + 10); columnCount = [Math]::Max($grid.columnCount, $t.Cols + 2) } }
                    fields = "gridProperties.rowCount,gridProperties.columnCount" } }
            }
        }
    }
    if ($isNew) {
        foreach ($title in @($existing.Keys)) {
            if ($titles -notcontains $title) { $requests += @{ deleteSheet = @{ sheetId = $existing[$title].sheetId } } }
        }
    }
    if ($requests.Count -gt 0) { Invoke-GoogleApi POST "$base`:batchUpdate" @{ requests = $requests } | Out-Null }

    # 2. Values
    Invoke-GoogleApi POST "$base/values:batchClear" @{ ranges = @($titles | ForEach-Object { Format-A1Sheet $_ }) } | Out-Null
    $data = @()
    foreach ($t in $tabs) {
        if ($t.Rows.Count -gt 0) {
            $data += @{ range = "$(Format-A1Sheet $t.Title)!A1"; majorDimension = "ROWS"; values = $t.Rows }
        }
    }
    if ($data.Count -gt 0) {
        Invoke-GoogleApi POST "$base/values:batchUpdate" @{ valueInputOption = "RAW"; data = $data } | Out-Null
    }

    # 3. Styling & Presentation
    $existing = Get-SheetProperties $base
    $formatRequests = @()

    for ($i = 0; $i -lt $tabs.Count; $i++) {
        $t = $tabs[$i]
        $sheetId = $existing[$t.Title].sheetId

        # Order tabs
        $formatRequests += @{ updateSheetProperties = @{
            properties = @{ sheetId = $sheetId; index = $i }
            fields = "index" } }

        if ($t.Type -eq "dashboard") {
            # Dashboard formatting
            # Header title (row 0)
            $formatRequests += @{ repeatCell = @{
                range = @{ sheetId = $sheetId; startRowIndex = 0; endRowIndex = 1; startColumnIndex = 0; endColumnIndex = 5 }
                cell = @{ userEnteredFormat = @{
                    backgroundColor = @{ red = 0.08; green = 0.15; blue = 0.28 }
                    textFormat = @{ bold = $true; fontSize = 14; foregroundColor = @{ red = 1; green = 1; blue = 1 } }
                } }
                fields = "userEnteredFormat.backgroundColor,userEnteredFormat.textFormat"
            } }
            # Section headers: rows 4, 10, 18, 25
            foreach ($rIdx in @(4, 10, 18, 25)) {
                $formatRequests += @{ repeatCell = @{
                    range = @{ sheetId = $sheetId; startRowIndex = $rIdx; endRowIndex = ($rIdx + 1); startColumnIndex = 0; endColumnIndex = 5 }
                    cell = @{ userEnteredFormat = @{
                        backgroundColor = @{ red = 0.2; green = 0.27; blue = 0.38 }
                        textFormat = @{ bold = $true; foregroundColor = @{ red = 1; green = 1; blue = 1 } }
                    } }
                    fields = "userEnteredFormat.backgroundColor,userEnteredFormat.textFormat"
                } }
            }
            # Resize dashboard columns
            $colWidths = @(220, 120, 120, 380, 150)
            for ($c = 0; $c -lt $colWidths.Count; $c++) {
                $formatRequests += @{ updateDimensionProperties = @{
                    range = @{ sheetId = $sheetId; dimension = "COLUMNS"; startIndex = $c; endIndex = ($c + 1) }
                    properties = @{ pixelSize = $colWidths[$c] }
                    fields = "pixelSize" } }
            }
        } else {
            # Data tabs formatting
            # Freeze header row
            $formatRequests += @{ updateSheetProperties = @{
                properties = @{ sheetId = $sheetId; gridProperties = @{ frozenRowCount = 1 } }
                fields = "gridProperties.frozenRowCount" } }

            # Header style: dark slate header with bold white text
            $formatRequests += @{ repeatCell = @{
                range = @{ sheetId = $sheetId; startRowIndex = 0; endRowIndex = 1; startColumnIndex = 0; endColumnIndex = $t.Cols }
                cell = @{ userEnteredFormat = @{
                    backgroundColor = @{ red = 0.12; green = 0.16; blue = 0.24 }
                    textFormat = @{ bold = $true; foregroundColor = @{ red = 1; green = 1; blue = 1 } }
                    verticalAlignment = "MIDDLE"
                } }
                fields = "userEnteredFormat.backgroundColor,userEnteredFormat.textFormat,userEnteredFormat.verticalAlignment"
            } }

            # Auto resize all columns
            $formatRequests += @{ autoResizeDimensions = @{ dimensions = @{
                sheetId = $sheetId; dimension = "COLUMNS"; startIndex = 0; endIndex = $t.Cols } } }

            # Specific column widths & wrapping for text columns
            if ($t.Rows.Count -gt 0) {
                $headerList = [object[]]$t.Rows[0]
                for ($col = 0; $col -lt $headerList.Length; $col++) {
                    $h = "$($headerList[$col])"
                    $targetWidth = 0
                    $shouldWrap = $false

                    if ($h -match "Rules Text") { $targetWidth = 400; $shouldWrap = $true }
                    elseif ($h -match "Previous Rules Text") { $targetWidth = 340; $shouldWrap = $true }
                    elseif ($h -match "Flag|Action") { $targetWidth = 320; $shouldWrap = $true }
                    elseif ($h -match "Note") { $targetWidth = 280; $shouldWrap = $true }
                    elseif ($h -match "Card Title") { $targetWidth = 190 }
                    elseif ($h -match "Priority|Status") { $targetWidth = 140 }

                    if ($targetWidth -gt 0) {
                        $formatRequests += @{ updateDimensionProperties = @{
                            range = @{ sheetId = $sheetId; dimension = "COLUMNS"; startIndex = $col; endIndex = ($col + 1) }
                            properties = @{ pixelSize = $targetWidth }
                            fields = "pixelSize" } }
                    }
                    if ($shouldWrap) {
                        $formatRequests += @{ repeatCell = @{
                            range = @{ sheetId = $sheetId; startColumnIndex = $col; endColumnIndex = ($col + 1) }
                            cell = @{ userEnteredFormat = @{ wrapStrategy = "WRAP" } }
                            fields = "userEnteredFormat.wrapStrategy" } }
                    }
                }
            }
        }
    }

    if ($formatRequests.Count -gt 0) {
        Invoke-GoogleApi POST "$base`:batchUpdate" @{ requests = $formatRequests } | Out-Null
    }
}

# -----------------------------------------------------------------------------
# Google Doc Publisher (Multipart HTML Upload)
# -----------------------------------------------------------------------------
function Generate-HtmlReport {
    $sb = New-Object System.Text.StringBuilder
    [void]$sb.AppendLine("<!DOCTYPE html><html><head><meta charset='utf-8'><title>Card Town - Card Standardization Audit Report</title>")
    [void]$sb.AppendLine("<style>")
    [void]$sb.AppendLine("body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; color: #1e293b; line-height: 1.5; margin: 24px; }")
    [void]$sb.AppendLine("h1 { color: #0f172a; border-bottom: 2px solid #3b82f6; padding-bottom: 8px; font-size: 24px; }")
    [void]$sb.AppendLine("h2 { color: #1e293b; margin-top: 28px; border-bottom: 1px solid #cbd5e1; padding-bottom: 6px; font-size: 18px; }")
    [void]$sb.AppendLine("h3 { color: #334155; margin-top: 20px; font-size: 15px; }")
    [void]$sb.AppendLine(".meta { color: #64748b; font-size: 13px; margin-bottom: 20px; }")
    [void]$sb.AppendLine(".kpi-table { width: 100%; border-collapse: collapse; margin-bottom: 24px; }")
    [void]$sb.AppendLine(".kpi-table th { background-color: #0f172a; color: white; padding: 10px 12px; text-align: left; font-size: 13px; }")
    [void]$sb.AppendLine(".kpi-table td { border: 1px solid #e2e8f0; padding: 10px 12px; font-size: 13px; }")
    [void]$sb.AppendLine(".kpi-table tr:nth-child(even) { background-color: #f8fafc; }")
    [void]$sb.AppendLine(".badge-high { background-color: #fee2e2; color: #991b1b; padding: 3px 8px; border-radius: 4px; font-weight: bold; font-size: 11px; }")
    [void]$sb.AppendLine(".badge-std { background-color: #fef3c7; color: #92400e; padding: 3px 8px; border-radius: 4px; font-weight: bold; font-size: 11px; }")
    [void]$sb.AppendLine(".card-table { width: 100%; border-collapse: collapse; margin-top: 12px; font-size: 12px; }")
    [void]$sb.AppendLine(".card-table th { background-color: #1e293b; color: white; padding: 8px 10px; text-align: left; }")
    [void]$sb.AppendLine(".card-table td { border: 1px solid #cbd5e1; padding: 8px 10px; vertical-align: top; }")
    [void]$sb.AppendLine(".card-table tr:nth-child(even) { background-color: #f8fafc; }")
    [void]$sb.AppendLine(".rules-text { font-family: Consolas, monospace; font-size: 11px; white-space: pre-wrap; }")
    [void]$sb.AppendLine(".flag-box { background-color: #fffbeb; border-left: 4px solid #f59e0b; padding: 12px; margin: 12px 0; font-size: 13px; }")
    [void]$sb.AppendLine("</style></head><body>")

    [void]$sb.AppendLine("<h1>CARD TOWN: CARD STANDARDIZATION AUDIT REPORT</h1>")
    [void]$sb.AppendLine("<div class='meta'><strong>Scope:</strong> Complete audit of all 167 cards across 10 balance sheets | <strong>Companion FRD:</strong> Card Data Pipeline FRD (Domain 7) | <strong>Standard:</strong> Magic: The Gathering (MTG) syntax & Target Enums</div>")

    # Executive Summary & KPIs
    [void]$sb.AppendLine("<h2>1. Executive Summary & Audit Metrics</h2>")
    [void]$sb.AppendLine("<p>This audit standardizes all card names, subtypes, deployment zones, and rules text across Card Town's balance sheets. Punctuation, sentence capitalization, canonical ability keywords, and timing preambles are aligned with engine requirements (<code>FR-TXT-001</code> through <code>FR-TXT-006</code>).</p>")
    
    [void]$sb.AppendLine("<table class='kpi-table'>")
    [void]$sb.AppendLine("<tr><th>Audit Metric</th><th>Count</th><th>Percentage</th><th>Description</th></tr>")
    [void]$sb.AppendLine("<tr><td><strong>Total Cards Audited</strong></td><td>$totalCards</td><td>100.0%</td><td>Complete coverage across all 10 CSV balance sheets</td></tr>")
    [void]$sb.AppendLine("<tr><td><strong>Rules Text Standardized</strong></td><td>$changedCards</td><td>$('{0:P1}' -f ($changedCards / $totalCards))</td><td>Capitalization, punctuation, zone names, and keyword cleanup</td></tr>")
    [void]$sb.AppendLine("<tr><td><strong>Flagged for Designer Review</strong></td><td>$flaggedCards</td><td>$('{0:P1}' -f ($flaggedCards / $totalCards))</td><td>Ambiguities, rounding rules, costs, and gameplay decisions requiring sign-off</td></tr>")
    [void]$sb.AppendLine("<tr><td><strong>Rules Text Unchanged</strong></td><td>$unchangedCards</td><td>$('{0:P1}' -f ($unchangedCards / $totalCards))</td><td>Cards already strictly conforming to grammar guidelines</td></tr>")
    [void]$sb.AppendLine("</table>")

    # High Priority Items
    [void]$sb.AppendLine("<h2>2. High-Priority Designer Review Items (Domain 7.1)</h2>")
    [void]$sb.AppendLine("<p>The following 9 items involve gameplay semantics or spelling corrections requiring formal design confirmation:</p>")

    $highItems = @(
        @{ Title = "Animoceps Rager"; Faction = "ANIMUS"; Type = "HEADQUARTERS"; Prev = "When this HQ is dealt damage, scratch once. This effect happens once per turn."; Std = "Whenever this Headquarters is dealt damage, Scratch 1. This ability triggers only once each turn."; Note = "Formalized per-turn trigger frequency limit (maxPerTurn: 1; FR-ABL-008)." },
        @{ Title = "Funeral Fungus"; Faction = "ANIMUS"; Type = "MANEUVER"; Prev = "exhaust chosen ready allied unit and return 1 unit card from your discard pile to your hand."; Std = "As an additional cost to enact this maneuver, exhaust a readied allied unit. Return chosen unit card from your discard pile to your hand."; Note = "Templated exhaust as additional casting cost rather than spell effect (FR-ABL-009) to eliminate fizzle ambiguity." },
        @{ Title = "Rift"; Faction = "ANIMUS"; Type = "HEADQUARTERS"; Prev = "discard a card, exhaust: scratch 2, deal 5 damage to this asset"; Std = "Discard a card, exhaust: Scratch 2. Deal 5 damage to this Headquarters."; Note = "Confirmed self-damage interaction on Headquarters; Headquarters is eligible target for damage mitigation (FR-REP-004)." },
        @{ Title = "Mutation Swell"; Faction = "ANIMUS"; Type = "MANEUVER"; Prev = "kill up to 4 allied scratch company units. for every 2 units killed, deploy 1 unit from your discard pile for free."; Std = "Kill up to 4 allied Scratch Company units. For every 2 units killed this way, deploy a unit card from your discard pile for free."; Note = "Clarified MTG causality clause ('killed this way'). Designer should confirm rounding if odd units killed (FR-ABL-007)." },
        @{ Title = "Steel Rain"; Faction = "PATH"; Type = "MANEUVER"; Prev = "discard up to 3 cards. For each card discarded this way, deal 2 damage to chosen asset. then, if you have no cards left in your hand, draw 2 cards."; Std = "Discard up to 3 cards. For each card discarded this way, deal 2 damage to chosen asset. Then, if you have no cards in hand, draw 2 cards."; Note = "Confirm if 'chosen asset' is selected once or independently per card discarded (FR-ABL-011)." },
        @{ Title = "Macrocannon Barrage"; Faction = "SOULLESS"; Type = "MANEUVER"; Prev = "deal 5 damage distributed among up to 5 chosen units"; Std = "Distribute 5 damage among up to 5 chosen units."; Note = "Standardized as distributed damage pool (DISTRIBUTED_POOL; FR-ABL-011)." },
        @{ Title = "Reverberating Terror"; Faction = "CONCORDIA"; Type = "MANEUVER"; Prev = "up to 2 chosen units and structures get -1 power until end of turn."; Std = "Up to 2 chosen units and structures get -1 power until end of turn."; Note = "Title spelling corrected ('reverbating terror' -> 'Reverberating Terror')." },
        @{ Title = "Devilhorn Guerrillas"; Faction = "PATH"; Type = "UNIT"; Prev = "when this unit deploys, if it was deployed to the main theater, deal 2 damage to chosen opposing unit."; Std = "When this unit deploys, if it was deployed to the Main Theater, deal 2 damage to chosen opposing unit."; Note = "Title spelling corrected ('Devilhorn Guerillas' -> 'Devilhorn Guerrillas')." },
        @{ Title = "All-Out War"; Faction = "SOULLESS"; Type = "MANEUVER"; Prev = "Discard your hand, then: - chosen allied asset gains power... - scratch a number of times... - regroup all..."; Std = "Discard your hand. Then: - Chosen allied asset gets power equal to the number of cards discarded this way until end of turn. - Scratch X, where X is the number of cards discarded this way. - Regroup all Scratch Company units in your Reserves."; Note = "Title hyphenated ('All-Out War'); modal list templating standardized (FR-ABL-010)." }
    )

    [void]$sb.AppendLine("<table class='card-table'>")
    [void]$sb.AppendLine("<tr><th>Card</th><th>Faction / Type</th><th>Previous Rules Text</th><th>Standardized Rules Text</th><th>Design Call / Rationale</th></tr>")
    foreach ($item in $highItems) {
        [void]$sb.AppendLine("<tr>")
        [void]$sb.AppendLine("<td><strong>$($item.Title)</strong><br><span class='badge-high'>HIGH PRIORITY</span></td>")
        [void]$sb.AppendLine("<td>$($item.Faction)<br><small>$($item.Type)</small></td>")
        [void]$sb.AppendLine("<td class='rules-text'>$([System.Web.HttpUtility]::HtmlEncode($item.Prev))</td>")
        [void]$sb.AppendLine("<td class='rules-text'><strong>$([System.Web.HttpUtility]::HtmlEncode($item.Std))</strong></td>")
        [void]$sb.AppendLine("<td>$([System.Web.HttpUtility]::HtmlEncode($item.Note))</td>")
        [void]$sb.AppendLine("</tr>")
    }
    [void]$sb.AppendLine("</table>")

    # Master Review Register (All 52 flagged cards)
    [void]$sb.AppendLine("<h2>3. Master Review Register (All 52 Flagged Cards)</h2>")
    [void]$sb.AppendLine("<p>All cards requiring designer confirmation or notable grammatical revisions:</p>")

    [void]$sb.AppendLine("<table class='card-table'>")
    [void]$sb.AppendLine("<tr><th>Card</th><th>Faction / Type</th><th>Designer Review Flag</th><th>Standardization Note</th><th>Standardized Rules Text</th><th>Previous Rules Text</th></tr>")

    $allFlaggedCards = @($highCards) + @($standardCards)
    foreach ($c in $allFlaggedCards) {
        $badgeClass = if ($c.StandardizedName -in $highPriorityNames) { "badge-high" } else { "badge-std" }
        $badgeText = if ($c.StandardizedName -in $highPriorityNames) { "HIGH PRIORITY" } else { "REVIEW" }

        [void]$sb.AppendLine("<tr>")
        [void]$sb.AppendLine("<td><strong>$([System.Web.HttpUtility]::HtmlEncode($c.StandardizedName))</strong><br><span class='$badgeClass'>$badgeText</span></td>")
        [void]$sb.AppendLine("<td>$($c.Faction)<br><small>$($c.CardType) / $($c.Rarity)</small></td>")
        [void]$sb.AppendLine("<td><strong>$([System.Web.HttpUtility]::HtmlEncode($c.Flag))</strong></td>")
        [void]$sb.AppendLine("<td>$([System.Web.HttpUtility]::HtmlEncode($c.Note))</td>")
        [void]$sb.AppendLine("<td class='rules-text'>$([System.Web.HttpUtility]::HtmlEncode($c.StandardizedText))</td>")
        [void]$sb.AppendLine("<td class='rules-text'>$([System.Web.HttpUtility]::HtmlEncode($c.OriginalText))</td>")
        [void]$sb.AppendLine("</tr>")
    }
    [void]$sb.AppendLine("</table>")

    # Templating Grammar Standards
    [void]$sb.AppendLine("<h2>4. MTG-Aligned Rules Templating Grammar Standards</h2>")
    [void]$sb.AppendLine("<ul>")
    [void]$sb.AppendLine("<li><strong>Trigger Preambles (FR-TXT-002):</strong> <code>When [event]</code> for one-time triggers (e.g. deployment); <code>Whenever [event]</code> for recurring triggers (combat, damage); <code>At [phase]</code> for phase/turn boundaries.</li>")
    [void]$sb.AppendLine("<li><strong>Cost Notation (FR-TXT-003):</strong> <code>[Cost]: [Effect].</code> Comma-separated costs ending with colon, effect sentence capitalized and terminated with period.</li>")
    [void]$sb.AppendLine("<li><strong>Targeting Language (FR-TXT-004):</strong> Canonical <code>chosen</code> replaces informal targeting (e.g. <code>chosen opposing unit</code>). Multi-target disambiguation uses <code>another chosen asset</code>.</li>")
    [void]$sb.AppendLine("<li><strong>Zone & Entity Capitalization:</strong> Capitalized <code>Main Theater</code>, <code>Reserves</code>, <code>Discard Pile</code>, <code>Deck</code>, <code>Headquarters</code>.</li>")
    [void]$sb.AppendLine("<li><strong>Keyword Formatting:</strong> Reminder text removed (e.g. <code>Aggressive 2</code>, <code>Resistant 1</code>, <code>Ethereal</code>, <code>Ward</code>, <code>Deploys Exhausted</code>).</li>")
    [void]$sb.AppendLine("<li><strong>Replacement & Causality Syntax:</strong> <code>If [event] would occur, [effect] instead</code> (FR-REP-001); causality clauses formatted as <code>killed this way</code> or <code>discarded this way</code>.</li>")
    [void]$sb.AppendLine("</ul>")

    [void]$sb.AppendLine("</body></html>")
    return $sb.ToString()
}

function Generate-MarkdownReport {
    $sb = New-Object System.Text.StringBuilder
    [void]$sb.AppendLine("# Card Town: Card Standardization Audit Report")
    [void]$sb.AppendLine("")
    [void]$sb.AppendLine("> **Scope**: Complete audit of all 167 cards across 10 balance sheets  ")
    [void]$sb.AppendLine("> **Companion FRD**: [Card Data Pipeline FRD](file:///docs/frd/02_Card_Data_Pipeline_FRD.md) (Domain 7)  ")
    [void]$sb.AppendLine("> **Grammar Standard**: Magic: The Gathering (MTG) syntax & Zero-Aliasing Target Enums (`FR-DAT-006`)")
    [void]$sb.AppendLine("")
    [void]$sb.AppendLine("---")
    [void]$sb.AppendLine("")
    [void]$sb.AppendLine("## 1. Executive Summary & Audit Metrics")
    [void]$sb.AppendLine("")
    [void]$sb.AppendLine("| Audit Metric | Count | Percentage | Description |")
    [void]$sb.AppendLine("| :--- | :--- | :--- | :--- |")
    [void]$sb.AppendLine("| **Total Cards Audited** | **$totalCards** | **100.0%** | All cards across all 10 balance CSV sheets |")
    [void]$sb.AppendLine("| **Rules Text Standardized** | **$changedCards** | **$('{0:P1}' -f ($changedCards / $totalCards))** | MTG capitalization, punctuation, zones, keywords |")
    [void]$sb.AppendLine("| **Flagged for Designer Review** | **$flaggedCards** | **$('{0:P1}' -f ($flaggedCards / $totalCards))** | Ambiguities, rounding rules, costs, requiring sign-off |")
    [void]$sb.AppendLine("| **Rules Text Unchanged** | **$unchangedCards** | **$('{0:P1}' -f ($unchangedCards / $totalCards))** | Cards already conforming to grammar guidelines |")
    [void]$sb.AppendLine("")
    [void]$sb.AppendLine("### Faction Breakdown")
    [void]$sb.AppendLine("")
    [void]$sb.AppendLine("| Faction | Total Cards | Maneuvers | Units & Structures | Headquarters |")
    [void]$sb.AppendLine("| :--- | :--- | :--- | :--- | :--- |")
    foreach ($f in @("ANIMUS", "CONCORDIA", "PATH", "SOULLESS")) {
        $fCards = $audit | Where-Object { $_.Faction -eq $f }
        $man = ($fCards | Where-Object { $_.CardType -eq "MANEUVER" }).Count
        $us = ($fCards | Where-Object { $_.CardType -in @("UNIT", "STRUCTURE") }).Count
        $hq = ($fCards | Where-Object { $_.CardType -eq "HEADQUARTERS" }).Count
        [void]$sb.AppendLine("| **$f** | $($fCards.Count) | $man | $us | $hq |")
    }
    $cCards = @($audit | Where-Object { $_.Faction -in @("COMMON", "NEUTRAL") })
    [void]$sb.AppendLine("| **COMMON / NEUTRAL** | $($cCards.Count) | 0 | $($cCards.Count) | 0 |")
    [void]$sb.AppendLine("")
    [void]$sb.AppendLine("---")
    [void]$sb.AppendLine("")
    [void]$sb.AppendLine("## 2. High-Priority Designer Review Items (Domain 7.1)")
    [void]$sb.AppendLine("")
    [void]$sb.AppendLine("| Card Title | Faction / Type | Previous Rules Text | Standardized Rules Text | Designer Review Flag / Rationale | Companion Requirement |")
    [void]$sb.AppendLine("| :--- | :--- | :--- | :--- | :--- | :--- |")
    [void]$sb.AppendLine("| **Animoceps Rager** | ANIMUS HQ | `When this HQ is dealt damage, scratch once. This effect happens once per turn.` | `Whenever this Headquarters is dealt damage, Scratch 1. This ability triggers only once each turn.` | Formalized per-turn limit syntax (`maxPerTurn: 1`). | `FR-ABL-008` |")
    [void]$sb.AppendLine("| **Funeral Fungus** | ANIMUS Maneuver | `exhaust chosen ready allied unit and return 1 unit card from your discard pile to your hand.` | `As an additional cost to enact this maneuver, exhaust a readied allied unit. Return chosen unit card from your discard pile to your hand.` | Templated exhaust as additional enactment cost rather than spell effect. | `FR-ABL-009` |")
    [void]$sb.AppendLine("| **Rift** | ANIMUS HQ | `discard a card, exhaust: scratch 2, deal 5 damage to this asset` | `Discard a card, exhaust: Scratch 2. Deal 5 damage to this Headquarters.` | Confirmed Headquarters self-damage keyword interaction (Unbowed). | `FR-REP-004` |")
    [void]$sb.AppendLine("| **Mutation Swell** | ANIMUS Maneuver | `kill up to 4 allied scratch company units. for every 2 units killed, deploy 1 unit from your discard pile for free.` | `Kill up to 4 allied Scratch Company units. For every 2 units killed this way, deploy a unit card from your discard pile for free.` | Clarified causality ('killed this way'); confirm odd-number rounding. | `FR-ABL-007` |")
    [void]$sb.AppendLine("| **Steel Rain** | PATH Maneuver | `discard up to 3 cards. For each card discarded this way, deal 2 damage to chosen asset. then, if you have no cards left in your hand, draw 2 cards.` | `Discard up to 3 cards. For each card discarded this way, deal 2 damage to chosen asset. Then, if you have no cards in hand, draw 2 cards.` | Confirm if 'chosen asset' is selected once or independently per card discarded. | `FR-ABL-011` |")
    [void]$sb.AppendLine("| **Macrocannon Barrage** | SOULLESS Maneuver | `deal 5 damage distributed among up to 5 chosen units` | `Distribute 5 damage among up to 5 chosen units.` | Templated as distributed damage pool (`DISTRIBUTED_POOL`). | `FR-ABL-011` |")
    [void]$sb.AppendLine("| **Reverberating Terror** | CONCORDIA Maneuver | `up to 2 chosen units and structures get -1 power until end of turn.` | `Up to 2 chosen units and structures get -1 power until end of turn.` | Title spelling corrected ('Reverberating Terror'). | `FR-TXT-001` |")
    [void]$sb.AppendLine("| **Devilhorn Guerrillas** | PATH Unit | `when this unit deploys, if it was deployed to the main theater, deal 2 damage to chosen opposing unit.` | `When this unit deploys, if it was deployed to the Main Theater, deal 2 damage to chosen opposing unit.` | Title spelling corrected ('Devilhorn Guerrillas'). | `FR-TXT-001` |")
    [void]$sb.AppendLine("| **All-Out War** | SOULLESS Maneuver | `Discard your hand, then: - chosen allied asset gains power...` | `Discard your hand. Then: - Chosen allied asset gets power equal to the number of cards discarded this way until end of turn. - Scratch X, where X is the number of cards discarded this way. - Regroup all Scratch Company units in your Reserves.` | Title hyphenated ('All-Out War'); list templating standardized. | `FR-ABL-010` |")
    [void]$sb.AppendLine("")
    [void]$sb.AppendLine("---")
    [void]$sb.AppendLine("")
    [void]$sb.AppendLine("## 3. Master Review Register (All 52 Flagged Cards)")
    [void]$sb.AppendLine("")
    [void]$sb.AppendLine("| Card Title | Faction | Type | Rarity | Designer Review Flag | Standardization Note | Standardized Rules Text | Previous Rules Text |")
    [void]$sb.AppendLine("| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |")

    $allFlaggedCards = @($highCards) + @($standardCards)
    foreach ($c in $allFlaggedCards) {
        $prio = if ($c.StandardizedName -in $highPriorityNames) { "**[HIGH]** " } else { "" }
        $stdClean = $c.StandardizedText -replace "`r`n", "<br>" -replace "`n", "<br>"
        $prevClean = $c.OriginalText -replace "`r`n", "<br>" -replace "`n", "<br>"
        [void]$sb.AppendLine("| $prio**$($c.StandardizedName)** | $($c.Faction) | $($c.CardType) | $($c.Rarity) | $($c.Flag) | $($c.Note) | $stdClean | $prevClean |")
    }

    return $sb.ToString()
}

function Publish-AuditGoogleDoc([string]$folderId, [string]$docName, [string]$htmlContent) {
    # Check if doc exists in destination folder
    $found = Find-DriveFiles "'$folderId' in parents and name = '$(Format-DriveQueryString $docName)' and mimeType = 'application/vnd.google-apps.document' and trashed = false"

    $boundary = "-------314159265358979323846"
    $delimiter = "`r`n--$boundary`r`n"
    $closeDelimiter = "`r`n--$boundary--`r`n"

    if ($found.Count -ge 1) {
        $fileId = $found[0].id
        # Update existing document content via uploadType=media
        $bytes = [Text.Encoding]::UTF8.GetBytes($htmlContent)
        $resp = Invoke-RestMethod -Method Patch -Uri "https://www.googleapis.com/upload/drive/v3/files/$fileId`?uploadType=media" `
            -Headers @{ Authorization = "Bearer $script:AccessToken" } `
            -ContentType "text/html; charset=utf-8" `
            -Body $bytes
        return @{ Id = $fileId; Created = $false }
    } else {
        # Create new document via multipart upload with mimeType conversion
        $metadata = @{
            name     = $docName
            mimeType = "application/vnd.google-apps.document"
            parents  = @($folderId)
        } | ConvertTo-Json -Compress

        $body = $delimiter +
            "Content-Type: application/json; charset=UTF-8`r`n`r`n" +
            $metadata +
            $delimiter +
            "Content-Type: text/html; charset=UTF-8`r`n`r`n" +
            $htmlContent +
            $closeDelimiter

        $bytes = [Text.Encoding]::UTF8.GetBytes($body)
        $resp = Invoke-RestMethod -Method Post -Uri "https://www.googleapis.com/upload/drive/v3/files?uploadType=multipart&fields=id,name,mimeType" `
            -Headers @{ Authorization = "Bearer $script:AccessToken" } `
            -ContentType "multipart/related; boundary=$boundary" `
            -Body $bytes

        return @{ Id = $resp.id; Created = $true }
    }
}

# -----------------------------------------------------------------------------
# Main Execution
# -----------------------------------------------------------------------------
Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "  CARD TOWN: PUBLISH STANDARDIZATION AUDIT TO GOOGLE DRIVE " -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "Audit JSON:     $auditJsonPath"
Write-Host "Destination:    $DestinationPath`n"
Write-Host "Total Cards:    $totalCards"
Write-Host "Standardized:   $changedCards"
Write-Host "Flagged Review: $flaggedCards"
Write-Host "Unchanged:      $unchangedCards`n"

# 1. Generate local synced reports
Write-Host "Generating human-readable report documents..." -ForegroundColor Yellow
$htmlReport = Generate-HtmlReport
$mdReport = Generate-MarkdownReport

# Write to local repository docs/
$repoDocPath = Join-Path $RepoRoot "docs\Card_Standardization_Audit_Report.md"
$mdReport | Set-Content -Path $repoDocPath -Encoding UTF8
Write-Host "  [OK] Saved to repository: $repoDocPath" -ForegroundColor Gray

# Write to Google Drive synced folder if it exists locally
if (Test-Path $DestinationPath) {
    $driveMdPath = Join-Path $DestinationPath "Card_Standardization_Audit_Report.md"
    $driveHtmlPath = Join-Path $DestinationPath "Card_Standardization_Audit_Report.html"
    $mdReport | Set-Content -Path $driveMdPath -Encoding UTF8
    $htmlReport | Set-Content -Path $driveHtmlPath -Encoding UTF8
    Write-Host "  [OK] Synced Markdown to Drive: $driveMdPath" -ForegroundColor Gray
    Write-Host "  [OK] Synced HTML to Drive:     $driveHtmlPath" -ForegroundColor Gray
} else {
    Write-Warning "Local Drive folder '$DestinationPath' is not mounted or not found. Local sync skipped."
}

if ($DryRun) {
    Write-Host "`nDry run complete: reports generated locally, nothing was sent to Google API." -ForegroundColor Green
    return
}

# 2. Authenticate to Google
Write-Host "`nConnecting to Google APIs..." -ForegroundColor Yellow
$script:AccessToken = Get-AccessToken
$folderId = Resolve-DriveFolderId $DestinationPath
Write-Host "  Resolved Card sheets folder ID: $folderId" -ForegroundColor Gray

# 3. Publish Native Google Sheet
$sheetName = "Card Standardization Audit"
Write-Host "`nPublishing native Google Sheet: '$sheetName'..." -ForegroundColor Yellow
$sheetObj = Get-OrCreateSpreadsheet $folderId $sheetName
Publish-AuditSpreadsheet $sheetObj.Id $sheetObj.Created $allTabs
$sheetVerb = if ($sheetObj.Created) { "Created" } else { "Updated" }
$sheetUrl = "https://docs.google.com/spreadsheets/d/$($sheetObj.Id)/edit"
Write-Host "  [$sheetVerb] Google Sheet -> $sheetUrl" -ForegroundColor Green

# 4. Publish Native Google Doc
$docName = "Card Standardization Audit Report"
Write-Host "`nPublishing native Google Doc: '$docName'..." -ForegroundColor Yellow
$docObj = Publish-AuditGoogleDoc $folderId $docName $htmlReport
$docVerb = if ($docObj.Created) { "Created" } else { "Updated" }
$docUrl = "https://docs.google.com/document/d/$($docObj.Id)/edit"
Write-Host "  [$docVerb] Google Doc   -> $docUrl" -ForegroundColor Green

Write-Host "`n==========================================================" -ForegroundColor Cyan
Write-Host "  AUDIT PUBLICATION COMPLETE!                              " -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "Interactive Google Sheet: $sheetUrl"
Write-Host "Document Report:          $docUrl"
Write-Host "Local Google Drive files: $DestinationPath`n"
