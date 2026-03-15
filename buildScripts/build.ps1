<#
.SYNOPSIS
    Builds the git-hunk Go binary.

.DESCRIPTION
    Runs go build ./... for the current module.
    Use -Release to produce a production binary in ./dist/.
    Use -CI to exit with a non-zero code on failure (suitable for GitHub Actions).

.PARAMETER Release
    Build an optimised production binary and place it in ./dist/git-hunk.

.PARAMETER CI
    Exit with code 1 on failure instead of returning normally.

.EXAMPLE
    ./buildScripts/build.ps1
    ./buildScripts/build.ps1 -Release -CI

.NOTES
    File Name  : build.ps1
    Author     : git-hunk contributors
    Prerequisite : Go 1.22+, PowerShell 7+
#>
[CmdletBinding()]
param(
    [switch]$Release,
    [switch]$CI
)

$ErrorActionPreference = 'Stop'

function Invoke-Build {
    if ($Release) {
        Write-Host 'Building production binary...' -ForegroundColor Cyan
        $null = New-Item -ItemType Directory -Force -Path './dist'
        $env:CGO_ENABLED = '0'
        $ldflags = '-s -w'
        go build -ldflags $ldflags -o './dist/git-hunk' './cmd/git-hunk/...'
        if ($LASTEXITCODE -ne 0) {
            throw "go build (release) failed with exit code $LASTEXITCODE"
        }
        Write-Host 'Release binary written to ./dist/git-hunk' -ForegroundColor Green
    }
    else {
        Write-Host 'Building all packages...' -ForegroundColor Cyan
        go build './...'
        if ($LASTEXITCODE -ne 0) {
            throw "go build failed with exit code $LASTEXITCODE"
        }
        Write-Host 'Build succeeded.' -ForegroundColor Green
    }
}

try {
    Invoke-Build
}
catch {
    Write-Host "Build failed: $_" -ForegroundColor Red
    if ($CI) { exit 1 }
    throw
}
