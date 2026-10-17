<#
.SYNOPSIS
    Merges the per-faction card CSVs into native Google Sheets (one tab per faction)
    inside the Card Town "Card sheets" Google Drive folder.

.DESCRIPTION
    data/cards/maneuvers-<faction>.csv            -> "maneuvers" Google Sheet
    data/cards/units-and-structures-<faction>.csv -> "units-and-structures" Google Sheet

    The tab name is the <faction> part of the file name, verbatim (e.g. "animus", "HQ").
    Each spreadsheet is created on the first run and updated in place on later runs,
    so its URL / Drive shortcut stays stable. Tabs that are not backed by a CSV are
    left untouched.

    Requires a one-time Google OAuth setup - see tools/sheets/README.md.

.EXAMPLE
    powershell -ExecutionPolicy Bypass -File tools\sheets\publish_card_sheets.ps1

.EXAMPLE
    # Parse the CSVs and show what would be published, without contacting Google.
    powershell -ExecutionPolicy Bypass -File tools\sheets\publish_card_sheets.ps1 -DryRun
#>
param(
    [string]$RepoRoot = (Resolve-Path "$PSScriptRoot\..\..").Path,
    # Local Google Drive for Desktop path of the destination folder. Only the part after
    # "My Drive\" is used, to locate the same folder through the Drive API.
    [string]$DestinationPath = "G:\My Drive\The End of the Manufacture\Card Town\Card sheets",
    [string]$ConfigDir = (Join-Path $env:APPDATA "animus-mundi"),
    [string]$ClientSecretPath,
    [switch]$DryRun,
    [switch]$ReAuth
)

$ErrorActionPreference = 'Stop'
Add-Type -AssemblyName Microsoft.VisualBasic
[Net.ServicePointManager]::SecurityProtocol = [Net.ServicePointManager]::SecurityProtocol -bor [Net.SecurityProtocolType]::Tls12

$cardsDir = Join-Path $RepoRoot "data\cards"
if (-not $ClientSecretPath) { $ClientSecretPath = Join-Path $ConfigDir "client_secret.json" }
$TokenPath = Join-Path $ConfigDir "google-refresh-token.dat"
# Full Drive scope is needed to find the existing "Card sheets" folder by name.
$Scopes = "https://www.googleapis.com/auth/drive"

$Workbooks = @(
    @{ Name = "maneuvers";            Prefix = "maneuvers-";            Expected = @("animus", "concordia", "path", "soulless") },
    @{ Name = "units-and-structures"; Prefix = "units-and-structures-"; Expected = @("animus", "concordia", "path", "soulless", "common", "HQ") }
)

# -----------------------------------------------------------------------------
# CSV loading
# -----------------------------------------------------------------------------
function ConvertTo-CellValue([string]$s) {
    # Send plain numbers as numbers so costs/stats are numeric in Sheets; everything
    # else is sent as literal text (RAW), so values like "+1" or "=" are never evaluated.
    if ($s -match '^-?\d+(\.\d+)?$' -and $s -notmatch '^-?0\d') {
        return [double]::Parse($s, [Globalization.CultureInfo]::InvariantCulture)
    }
    return $s
}

function Read-CsvRows([string]$path) {
    $parser = New-Object Microsoft.VisualBasic.FileIO.TextFieldParser($path, [Text.Encoding]::UTF8)
    try {
        $parser.TextFieldType = [Microsoft.VisualBasic.FileIO.FieldType]::Delimited
        $parser.SetDelimiters(",")
        $parser.HasFieldsEnclosedInQuotes = $true
        $parser.TrimWhiteSpace = $false
        $rows = New-Object System.Collections.Generic.List[object]
        while (-not $parser.EndOfData) {
            $fields = $parser.ReadFields()
            $row = New-Object object[] $fields.Length
            for ($i = 0; $i -lt $fields.Length; $i++) { $row[$i] = ConvertTo-CellValue $fields[$i] }
            $rows.Add($row)
        }
    } finally { $parser.Close() }

    # Drop trailing rows that are completely empty.
    while ($rows.Count -gt 0 -and -not ($rows[$rows.Count - 1] | Where-Object { "$_" -ne "" })) {
        $rows.RemoveAt($rows.Count - 1)
    }
    return , $rows.ToArray()
}

function Get-WorkbookTabs($wb) {
    $files = @(Get-ChildItem -Path $cardsDir -Filter "$($wb.Prefix)*.csv" -File)
    $byFaction = @{}
    foreach ($f in $files) { $byFaction[$f.BaseName.Substring($wb.Prefix.Length)] = $f }

    foreach ($e in $wb.Expected) {
        if (-not $byFaction.ContainsKey($e)) { Write-Warning "$($wb.Name): expected '$($wb.Prefix)$e.csv' was not found." }
    }

    # Expected factions first (in the defined order), then any extras alphabetically.
    $ordered = @($wb.Expected | Where-Object { $byFaction.ContainsKey($_) })
    $ordered += @($byFaction.Keys | Where-Object { $wb.Expected -notcontains $_ } | Sort-Object)

    $tabs = @()
    foreach ($faction in $ordered) {
        $file = $byFaction[$faction]
        $rows = Read-CsvRows $file.FullName
        $cols = 0
        foreach ($r in $rows) { if ($r.Length -gt $cols) { $cols = $r.Length } }
        $tabs += @{ Title = $file.BaseName.Substring($wb.Prefix.Length); File = $file.Name; Rows = $rows; Cols = [Math]::Max($cols, 1) }
    }
    return , $tabs
}

# -----------------------------------------------------------------------------
# Google OAuth (installed-app loopback flow with PKCE)
# -----------------------------------------------------------------------------
function Get-ClientConfig {
    if (-not (Test-Path $ClientSecretPath) -and (Test-Path "$ClientSecretPath.json")) {
        # Explorer hides known extensions, so "client_secret.json" typed there becomes "client_secret.json.json".
        $script:ClientSecretPath = "$ClientSecretPath.json"
    }
    if (-not (Test-Path $ClientSecretPath)) {
        $dir = Split-Path $ClientSecretPath -Parent
        $present = if (Test-Path $dir) { (@(Get-ChildItem $dir -File | ForEach-Object { $_.Name }) -join ", ") } else { "(folder does not exist)" }
        throw "Google OAuth client file not found at '$ClientSecretPath'. Folder contains: $present. See tools\sheets\README.md for the one-time setup."
    }
    $json = Get-Content $ClientSecretPath -Raw | ConvertFrom-Json
    if (-not $json.installed) {
        throw "'$ClientSecretPath' is not a 'Desktop app' OAuth client. Create a Desktop app client (see tools\sheets\README.md)."
    }
    return $json.installed
}

function ConvertTo-Base64Url([byte[]]$bytes) {
    return [Convert]::ToBase64String($bytes).TrimEnd('=').Replace('+', '-').Replace('/', '_')
}

function Save-RefreshToken([string]$token) {
    New-Item -ItemType Directory -Path $ConfigDir -Force | Out-Null
    # Encrypted with Windows DPAPI: only this Windows user on this machine can read it.
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
                $ctx.Response.StatusCode = 404; $ctx.Response.Close(); continue   # e.g. favicon
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
# Google API helpers
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
    $uri = "https://www.googleapis.com/drive/v3/files?q=" + [Uri]::EscapeDataString($query) + "&fields=files(id,name)&spaces=drive"
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
    if ($found.Count -gt 1) { Write-Warning "Multiple spreadsheets named '$name' in the destination folder; updating the first." }
    if ($found.Count -ge 1) { return @{ Id = $found[0].id; Created = $false } }

    $file = Invoke-GoogleApi POST "https://www.googleapis.com/drive/v3/files?fields=id" @{
        name     = $name
        mimeType = "application/vnd.google-apps.spreadsheet"
        parents  = @($folderId)
    }
    return @{ Id = $file.id; Created = $true }
}

function Get-SheetProperties([string]$base) {
    $map = @{}
    foreach ($s in @((Invoke-GoogleApi GET "$base`?fields=sheets.properties").sheets)) { $map[$s.properties.title] = $s.properties }
    return $map
}

function Format-A1Sheet([string]$title) { return "'" + $title.Replace("'", "''") + "'" }

function Publish-Workbook([string]$spreadsheetId, [bool]$isNew, $tabs) {
    $base = "https://sheets.googleapis.com/v4/spreadsheets/$spreadsheetId"
    $existing = Get-SheetProperties $base
    $titles = @($tabs | ForEach-Object { $_.Title })

    # 1. Structure: add missing tabs, grow undersized grids, drop the default tab of a new file.
    $requests = @()
    foreach ($t in $tabs) {
        $needRows = [Math]::Max($t.Rows.Count, 1)
        if (-not $existing.ContainsKey($t.Title)) {
            $requests += @{ addSheet = @{ properties = @{ title = $t.Title; gridProperties = @{
                rowCount = [Math]::Max($needRows, 1000); columnCount = [Math]::Max($t.Cols, 26) } } } }
        } else {
            $grid = $existing[$t.Title].gridProperties
            if ($grid.rowCount -lt $needRows -or $grid.columnCount -lt $t.Cols) {
                $requests += @{ updateSheetProperties = @{
                    properties = @{ sheetId = $existing[$t.Title].sheetId; gridProperties = @{
                        rowCount = [Math]::Max($grid.rowCount, $needRows); columnCount = [Math]::Max($grid.columnCount, $t.Cols) } }
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

    # 2. Values: clear each managed tab, then write the CSV contents verbatim.
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

    # 3. Presentation: tab order, frozen bold header, sized columns, wrapped "Text" column.
    $existing = Get-SheetProperties $base
    $requests = @()
    for ($i = 0; $i -lt $tabs.Count; $i++) {
        $t = $tabs[$i]
        $sheetId = $existing[$t.Title].sheetId
        $requests += @{ updateSheetProperties = @{
            properties = @{ sheetId = $sheetId; index = $i; gridProperties = @{ frozenRowCount = 1 } }
            fields = "index,gridProperties.frozenRowCount" } }
        $requests += @{ repeatCell = @{
            range = @{ sheetId = $sheetId; startRowIndex = 0; endRowIndex = 1 }
            cell = @{ userEnteredFormat = @{ textFormat = @{ bold = $true } } }
            fields = "userEnteredFormat.textFormat.bold" } }
        $requests += @{ autoResizeDimensions = @{ dimensions = @{
            sheetId = $sheetId; dimension = "COLUMNS"; startIndex = 0; endIndex = $t.Cols } } }

        if ($t.Rows.Count -gt 0) {
            $textCol = [Array]::FindIndex([object[]]$t.Rows[0], [Predicate[object]] { param($h) "$h" -eq "rulesText" -or "$h" -eq "Text" })
            if ($textCol -ge 0) {
                $requests += @{ updateDimensionProperties = @{
                    range = @{ sheetId = $sheetId; dimension = "COLUMNS"; startIndex = $textCol; endIndex = $textCol + 1 }
                    properties = @{ pixelSize = 400 }
                    fields = "pixelSize" } }
                $requests += @{ repeatCell = @{
                    range = @{ sheetId = $sheetId; startColumnIndex = $textCol; endColumnIndex = $textCol + 1 }
                    cell = @{ userEnteredFormat = @{ wrapStrategy = "WRAP" } }
                    fields = "userEnteredFormat.wrapStrategy" } }
            }
        }
    }
    Invoke-GoogleApi POST "$base`:batchUpdate" @{ requests = $requests } | Out-Null
}

# -----------------------------------------------------------------------------
# Main
# -----------------------------------------------------------------------------
Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "  CARD TOWN: PUBLISH CARD SHEETS TO GOOGLE SHEETS          " -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "Cards Directory: $cardsDir"
Write-Host "Destination:     $DestinationPath`n"

$plan = @()
foreach ($wb in $Workbooks) {
    $tabs = Get-WorkbookTabs $wb
    if ($tabs.Count -eq 0) { Write-Warning "$($wb.Name): no CSVs found matching '$($wb.Prefix)*.csv'; skipping."; continue }
    Write-Host "$($wb.Name)" -ForegroundColor Yellow
    foreach ($t in $tabs) {
        Write-Host ("  [{0}] <- {1} ({2} rows incl. header, {3} columns)" -f $t.Title, $t.File, $t.Rows.Count, $t.Cols) -ForegroundColor Gray
    }
    $plan += @{ Workbook = $wb; Tabs = $tabs }
}

if ($DryRun) {
    Write-Host "`nDry run: nothing was sent to Google." -ForegroundColor Green
    return
}

$script:AccessToken = Get-AccessToken
$folderId = Resolve-DriveFolderId $DestinationPath

Write-Host ""
foreach ($p in $plan) {
    $name = $p.Workbook.Name
    $sheet = Get-OrCreateSpreadsheet $folderId $name
    Publish-Workbook $sheet.Id $sheet.Created $p.Tabs
    $verb = if ($sheet.Created) { "Created" } else { "Updated" }
    Write-Host "[$verb] $name -> https://docs.google.com/spreadsheets/d/$($sheet.Id)/edit" -ForegroundColor Green
}

Write-Host "`nDone. Drive for Desktop will show them as '<name>.gsheet' in:`n  $DestinationPath" -ForegroundColor Cyan
