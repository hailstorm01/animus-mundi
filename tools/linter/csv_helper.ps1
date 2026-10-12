# Helper to format CSV row with quotes around every field
function Export-QuotedCsv {
    param(
        [array]$Rows,
        [string]$Path
    )
    $lines = [System.Collections.Generic.List[string]]::new()
    if ($Rows.Count -eq 0) { return }
    
    # Headers
    $headers = $Rows[0].PSObject.Properties | Where-Object { $_.MemberType -eq 'NoteProperty' } | ForEach-Object { $_.Name }
    $headerLine = ($headers | ForEach-Object { '"' + $_.Replace('"', '""') + '"' }) -join ','
    $lines.Add($headerLine)
    
    foreach ($r in $Rows) {
        $fields = [System.Collections.Generic.List[string]]::new()
        foreach ($h in $headers) {
            $val = if ($r.$h -ne $null) { [string]$r.$h } else { "" }
            # Normalize internal newlines to \n
            $val = $val.Replace("`r`n", "`n").Replace("`r", "`n")
            $fields.Add('"' + $val.Replace('"', '""') + '"')
        }
        $lines.Add(($fields -join ','))
    }
    
    # Write UTF8 without BOM (standard RFC-4180 / Git friendly)
    $utf8NoBom = New-Object System.Text.UTF8Encoding($false)
    [System.IO.File]::WriteAllText($Path, ($lines -join "`r`n") + "`r`n", $utf8NoBom)
}
