<#
.SYNOPSIS
    Checks for outdated Go modules and known vulnerabilities.

.DESCRIPTION
    Uses 'go list -u -m all' to detect outdated dependencies and
    govulncheck to audit for known CVEs. If govulncheck is not present
    it is installed automatically.
    Use -Outdated to show outdated modules only.
    Use -Audit to run vulnerability checks only.
    When neither switch is provided, both checks run.
    Use -CI to exit with a non-zero code on failure.

.PARAMETER Outdated
    Check for outdated modules using 'go list -u -m all'.

.PARAMETER Audit
    Check for vulnerabilities using govulncheck.

.PARAMETER CI
    Exit with code 1 on failure instead of returning normally.

.EXAMPLE
    ./buildScripts/check-deps.ps1
    ./buildScripts/check-deps.ps1 -Audit -CI

.NOTES
    File Name  : check-deps.ps1
    Author     : git-hunk contributors
    Prerequisite : Go 1.22+, PowerShell 7+
#>
[CmdletBinding()]
param(
    [switch]$Outdated,
    [switch]$Audit,
    [switch]$CI
)

$ErrorActionPreference = 'Stop'

$runAll = -not $Outdated -and -not $Audit

function Test-Outdated {
    Write-Host 'Checking for outdated modules...' -ForegroundColor Cyan
    $output = go list -u -m all 2>&1
    if ($LASTEXITCODE -ne 0) {
        throw "go list failed with exit code $LASTEXITCODE"
    }
    $outdatedLines = $output | Where-Object { $_ -match '\[' }
    if ($outdatedLines) {
        Write-Host 'Outdated modules found:' -ForegroundColor Yellow
        $outdatedLines | ForEach-Object { Write-Host "  $_" -ForegroundColor Yellow }
    }
    else {
        Write-Host 'All modules are up to date.' -ForegroundColor Green
    }
}

function Install-Govulncheck {
    Write-Host 'govulncheck not found — installing...' -ForegroundColor Yellow
    go install golang.org/x/vuln/cmd/govulncheck@latest
    if ($LASTEXITCODE -ne 0) {
        throw "Failed to install govulncheck"
    }
    Write-Host 'govulncheck installed.' -ForegroundColor Green
}

function Test-Vulnerabilities {
    $vulnCmd = Get-Command govulncheck -ErrorAction SilentlyContinue
    if (-not $vulnCmd) {
        Install-Govulncheck
    }

    Write-Host 'Running vulnerability audit...' -ForegroundColor Cyan
    govulncheck './...'
    if ($LASTEXITCODE -ne 0) {
        throw "govulncheck found vulnerabilities (exit code $LASTEXITCODE)"
    }
    Write-Host 'No vulnerabilities found.' -ForegroundColor Green
}

try {
    if ($runAll -or $Outdated) { Test-Outdated }
    if ($runAll -or $Audit) { Test-Vulnerabilities }
}
catch {
    Write-Host "Dependency check failed: $_" -ForegroundColor Red
    if ($CI) { exit 1 }
    throw
}
