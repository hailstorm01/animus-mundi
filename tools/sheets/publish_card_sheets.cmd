@echo off
rem Publishes data\cards CSVs to the Card Town Google Sheets. Extra args are passed through (e.g. -DryRun, -ReAuth).
powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0publish_card_sheets.ps1" %*
if "%~1"=="" pause
