<#
.SYNOPSIS
    Lints PowerShell scripts in buildScripts/ using PSScriptAnalyzer.

.DESCRIPTION
    Runs Invoke-ScriptAnalyzer against all *.ps1 files in buildScripts/.
    PSScriptAnalyzer is auto-installed from the PowerShell Gallery if not present.
    Settings are read from .PSScriptAnalyzerSettings.psd1 at the project root.
    Use -Fix to apply automatic fixes where possible.
    Use -CI to exit with a non-zero code when issues are found.

.PARAMETER Fix
    Apply PSScriptAnalyzer auto-fixes (SuggestedCorrections) where available.

.PARAMETER CI
    Exit with code 1 when lint issues are found.

.EXAMPLE
    ./buildScripts/lint-powershell.ps1
    ./buildScripts/lint-powershell.ps1 -Fix -CI

.NOTES
    File Name  : lint-powershell.ps1
    Author     : git-hunk contributors
    Prerequisite : PowerShell 7+
#>
[CmdletBinding()]
param(
    [switch]$Fix,
    [switch]$CI
)

$ErrorActionPreference = 'Stop'

function Install-PSScriptAnalyzer {
    Write-Host 'PSScriptAnalyzer not found — installing from PSGallery...' -ForegroundColor Yellow
    Install-Module -Name PSScriptAnalyzer -Scope CurrentUser -Force -AllowClobber
    Write-Host 'PSScriptAnalyzer installed.' -ForegroundColor Green
}

function Invoke-PSLint {
    if (-not (Get-Module -ListAvailable -Name PSScriptAnalyzer)) {
        Install-PSScriptAnalyzer
    }
    Import-Module PSScriptAnalyzer -ErrorAction Stop

    $settingsFile = Join-Path $PSScriptRoot '..' '.PSScriptAnalyzerSettings.psd1'
    $settingsFile = (Resolve-Path $settingsFile -ErrorAction SilentlyContinue)?.Path

    $scriptDir = $PSScriptRoot
    $scripts = Get-ChildItem -Path $scriptDir -Filter '*.ps1' -File

    Write-Host "Analysing $($scripts.Count) PowerShell script(s) in $scriptDir..." -ForegroundColor Cyan

    $results = @()
    foreach ($script in $scripts) {
        $params = @{ Path = $script.FullName }
        if ($settingsFile) { $params['Settings'] = $settingsFile }

        if ($Fix) {
            $params['Fix'] = $true
        }

        $findings = Invoke-ScriptAnalyzer @params
        $results += $findings
    }

    if ($results) {
        $results | Format-Table -AutoSize
        $errorCount = ($results | Where-Object { $_.Severity -eq 'Error' }).Count
        $warnCount  = ($results | Where-Object { $_.Severity -eq 'Warning' }).Count
        Write-Host "PSScriptAnalyzer: $errorCount error(s), $warnCount warning(s)." -ForegroundColor Yellow
        if ($errorCount -gt 0) {
            throw "PSScriptAnalyzer found $errorCount error(s)."
        }
    }
    else {
        Write-Host 'PSScriptAnalyzer: no issues found.' -ForegroundColor Green
    }
}

try {
    Invoke-PSLint
}
catch {
    Write-Host "PowerShell lint failed: $_" -ForegroundColor Red
    if ($CI) { exit 1 }
    throw
}
