<#
.SYNOPSIS
    Runs Pester 5 tests for the git-hunk build scripts.

.DESCRIPTION
    Discovers and runs all *.Tests.ps1 files in buildScripts/.
    Pester 5 is auto-installed from the PowerShell Gallery if not present.
    Use -Coverage to collect code coverage.
    Use -CI to exit with a non-zero code when tests fail.
    Use -TestType to filter: Unit, Integration, or All (default: All).

.PARAMETER Coverage
    Collect code coverage and output a JaCoCo-compatible report.

.PARAMETER CI
    Exit with code 1 when any tests fail.

.PARAMETER TestType
    Filter tests by type tag: Unit, Integration, or All.

.EXAMPLE
    ./buildScripts/run-pester.ps1
    ./buildScripts/run-pester.ps1 -CI -Coverage -TestType Unit

.NOTES
    File Name  : run-pester.ps1
    Author     : git-hunk contributors
    Prerequisite : PowerShell 7+
#>
[CmdletBinding()]
param(
    [switch]$Coverage,
    [switch]$CI,
    [ValidateSet('Unit', 'Integration', 'All')]
    [string]$TestType = 'All'
)

$ErrorActionPreference = 'Stop'

function Install-Pester {
    Write-Host 'Pester 5 not found — installing from PSGallery...' -ForegroundColor Yellow
    Install-Module -Name Pester -MinimumVersion 5.0.0 -Scope CurrentUser -Force -AllowClobber
    Write-Host 'Pester installed.' -ForegroundColor Green
}

function Invoke-PesterTests {
    $pesterModule = Get-Module -ListAvailable -Name Pester |
        Where-Object { $_.Version -ge [version]'5.0.0' } |
        Sort-Object Version -Descending |
        Select-Object -First 1

    if (-not $pesterModule) {
        Install-Pester
    }

    Import-Module Pester -MinimumVersion 5.0.0 -ErrorAction Stop

    $scriptDir = $PSScriptRoot

    # Discover test files.
    $testFiles = Get-ChildItem -Path $scriptDir -Filter '*.Tests.ps1' -File

    if ($TestType -ne 'All') {
        $testFiles = $testFiles | Where-Object { $_.Name -match $TestType }
    }

    if ($testFiles.Count -eq 0) {
        Write-Host 'No Pester test files found.' -ForegroundColor Yellow
        return
    }

    Write-Host "Running $($testFiles.Count) Pester test file(s) ($TestType)..." -ForegroundColor Cyan

    $config = New-PesterConfiguration
    $config.Run.Path = $testFiles.FullName
    $config.Output.Verbosity = 'Detailed'

    if ($Coverage) {
        $config.CodeCoverage.Enabled = $true
        $config.CodeCoverage.Path = (Get-ChildItem -Path $scriptDir -Filter '*.ps1' -Exclude '*.Tests.ps1').FullName
        $config.CodeCoverage.OutputPath = Join-Path $scriptDir '..' 'coverage-ps.xml'
        $config.CodeCoverage.OutputFormat = 'JaCoCo'
    }

    $result = Invoke-Pester -Configuration $config

    if ($result.FailedCount -gt 0) {
        throw "Pester: $($result.FailedCount) test(s) failed."
    }

    Write-Host "Pester: $($result.PassedCount) test(s) passed." -ForegroundColor Green
}

try {
    Invoke-PesterTests
}
catch {
    Write-Host "Pester run failed: $_" -ForegroundColor Red
    if ($CI) { exit 1 }
    throw
}
