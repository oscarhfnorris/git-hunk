<#
.SYNOPSIS
    Runs Go tests for git-hunk.

.DESCRIPTION
    Runs go test ./... for the current module.
    Use -Coverage to write a coverage profile to ./coverage.out.
    Use -CI to exit with a non-zero code on failure (suitable for GitHub Actions).

.PARAMETER Coverage
    Write coverage data to ./coverage.out and print a summary.

.PARAMETER CI
    Exit with code 1 on failure instead of returning normally.

.EXAMPLE
    ./buildScripts/test.ps1
    ./buildScripts/test.ps1 -Coverage -CI

.NOTES
    File Name  : test.ps1
    Author     : git-hunk contributors
    Prerequisite : Go 1.22+, PowerShell 7+
#>
[CmdletBinding()]
param(
    [switch]$Coverage,
    [switch]$CI
)

$ErrorActionPreference = 'Stop'

function Invoke-Tests {
    if ($Coverage) {
        Write-Host 'Running tests with coverage...' -ForegroundColor Cyan
        go test -coverprofile='coverage.out' -covermode=atomic './...'
        if ($LASTEXITCODE -ne 0) {
            throw "go test failed with exit code $LASTEXITCODE"
        }
        go tool cover -func='coverage.out'
        Write-Host 'Coverage profile written to coverage.out' -ForegroundColor Green
    }
    else {
        Write-Host 'Running tests...' -ForegroundColor Cyan
        go test './...'
        if ($LASTEXITCODE -ne 0) {
            throw "go test failed with exit code $LASTEXITCODE"
        }
    }
    Write-Host 'All tests passed.' -ForegroundColor Green
}

try {
    Invoke-Tests
}
catch {
    Write-Host "Tests failed: $_" -ForegroundColor Red
    if ($CI) { exit 1 }
    throw
}
