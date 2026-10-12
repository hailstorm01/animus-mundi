# Card Sheet Publisher

Merges the per-faction card CSVs in `data/cards/` into two native Google Sheets in
`G:\My Drive\The End of the Manufacture\Card Town\Card sheets`:

| Google Sheet           | Source CSVs                                   | Tabs                                                  |
| ---------------------- | --------------------------------------------- | ----------------------------------------------------- |
| `maneuvers`            | `maneuvers-<faction>.csv`                     | `animus`, `concordia`, `path`, `soulless`             |
| `units-and-structures` | `units-and-structures-<faction>.csv`          | `animus`, `concordia`, `path`, `soulless`, `common`, `HQ` |

- The tab name is the `<faction>` part of the file name, verbatim. A new CSV matching the pattern becomes a new tab automatically.
- The first run creates the sheets. Later runs overwrite each managed tab's contents **in place**, so links and the Drive shortcut stay the same. Tabs without a matching CSV are left alone.
- Numbers are stored as numbers. All other values are written as literal text, so nothing in a card's text is treated as a formula.
- Formatting: bold, frozen header row; auto-sized columns; the `Text` column wraps at 400 px.

## Usage

```powershell
# Publish (or double-click tools\sheets\publish_card_sheets.cmd)
powershell -ExecutionPolicy Bypass -File tools\sheets\publish_card_sheets.ps1

# Preview what would be published, without contacting Google
powershell -ExecutionPolicy Bypass -File tools\sheets\publish_card_sheets.ps1 -DryRun

# Force a fresh Google sign-in
powershell -ExecutionPolicy Bypass -File tools\sheets\publish_card_sheets.ps1 -ReAuth
```

Run the linter (`tools/linter/lint_cardtown.ps1`) before publishing.

## One-time Google setup

A script can't create Google Sheets by writing files into the `G:\` drive, because `.gsheet` files are just shortcuts. So this tool calls the Google Drive and Sheets APIs, which need an OAuth client of your own:

1. Open the [Google Cloud Console](https://console.cloud.google.com/) and create (or pick) a project.
2. **APIs & Services → Library**: enable **Google Drive API** and **Google Sheets API**.
3. **Google Auth Platform → Branding / Audience**: configure the consent screen as *External* and add your Google account under **Test users**.
   - Optional: click **Publish app**. While an app is in *Testing*, Google expires the sign-in every 7 days, and the script will open the browser again when that happens.
4. **Google Auth Platform → Clients → Create client**: choose application type **Desktop app**, then download the JSON.
5. Save the file as `%APPDATA%\animus-mundi\client_secret.json` (or pass `-ClientSecretPath <file>`).
6. Run the script. A browser window opens once for sign-in. Google will warn that the app is unverified; choose *Advanced → Go to (app)*, which is expected for a personal app.

The refresh token is stored in `%APPDATA%\animus-mundi\google-refresh-token.dat`, encrypted with Windows DPAPI (readable only by your Windows user). Never commit `client_secret.json`.

The tool requests the full `drive` scope. It needs this to find the existing `Card sheets` folder by name, and it only ever touches the `maneuvers` and `units-and-structures` spreadsheets in that folder.
