<#
.SYNOPSIS
    Lints Go source files using golangci-lint.

.DESCRIPTION
    Runs golangci-lint run. If golangci-lint is not present it is downloaded
    automatically. Use -Fix to apply auto-fixes. Use -CI to exit with a
    non-zero code on failure (suitable for GitHub Actions).

.PARAMETER Fix
    Pass --fix to golangci-lint to apply automatic fixes.

.PARAMETER CI
    Exit with code 1 on failure instead of returning normally.

.EXAMPLE
    ./buildScripts/lint.ps1
    ./buildScripts/lint.ps1 -Fix -CI

.NOTES
    File Name  : lint.ps1
    Author     : git-hunk contributors
    Prerequisite : Go 1.22+, PowerShell 7+
#>
[CmdletBinding()]
param(
    [switch]$Fix,
    [switch]$CI
)

$ErrorActionPreference = 'Stop'

function Install-GolangciLint {
    Write-Host 'golangci-lint not found — installing...' -ForegroundColor Yellow
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    if ($LASTEXITCODE -ne 0) {
        throw "Failed to install golangci-lint"
    }
    Write-Host 'golangci-lint installed.' -ForegroundColor Green
}

function Invoke-Lint {
    $lintCmd = Get-Command golangci-lint -ErrorAction SilentlyContinue
    if (-not $lintCmd) {
        Install-GolangciLint
    }

    $lintArgs = @('run', './...')
    if ($Fix) {
        $lintArgs += '--fix'
    }

    Write-Host "Running golangci-lint $($lintArgs -join ' ')..." -ForegroundColor Cyan
    & golangci-lint @lintArgs
    if ($LASTEXITCODE -ne 0) {
        throw "golangci-lint failed with exit code $LASTEXITCODE"
    }
    Write-Host 'Lint passed.' -ForegroundColor Green
}

try {
    Invoke-Lint
}
catch {
    Write-Host "Lint failed: $_" -ForegroundColor Red
    if ($CI) { exit 1 }
    throw
}
