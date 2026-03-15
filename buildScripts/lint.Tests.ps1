#Requires -Version 7
<#
.SYNOPSIS
    Pester tests for buildScripts/lint.ps1
#>

BeforeAll {
    $scriptPath = Join-Path $PSScriptRoot 'lint.ps1'
    $ast = [System.Management.Automation.Language.Parser]::ParseFile(
        $scriptPath, [ref]$null, [ref]$null)
    $functionDefs = $ast.FindAll(
        { param($node) $node -is [System.Management.Automation.Language.FunctionDefinitionAst] },
        $true)
    foreach ($fn in $functionDefs) {
        Invoke-Expression $fn.Extent.Text
    }
}

Describe 'lint.ps1' {
    Context 'Functions' {
        It 'Invoke-Lint is defined' {
            Get-Command Invoke-Lint -ErrorAction SilentlyContinue | Should -Not -BeNullOrEmpty
        }

        It 'Install-GolangciLint is defined' {
            Get-Command Install-GolangciLint -ErrorAction SilentlyContinue | Should -Not -BeNullOrEmpty
        }
    }

    Context 'Script structure' {
        It 'Script file exists' {
            Test-Path (Join-Path $PSScriptRoot 'lint.ps1') | Should -Be $true
        }

        It 'Has ErrorActionPreference Stop' {
            $content = Get-Content (Join-Path $PSScriptRoot 'lint.ps1') -Raw
            $content | Should -Match "\`$ErrorActionPreference\s*=\s*'Stop'"
        }

        It 'Has comment-based help' {
            $content = Get-Content (Join-Path $PSScriptRoot 'lint.ps1') -Raw
            $content | Should -Match '\.SYNOPSIS'
            $content | Should -Match '\.NOTES'
        }

        It 'Has -Fix switch' {
            $content = Get-Content (Join-Path $PSScriptRoot 'lint.ps1') -Raw
            $content | Should -Match '\[switch\]\$Fix'
        }

        It 'Has -CI switch' {
            $content = Get-Content (Join-Path $PSScriptRoot 'lint.ps1') -Raw
            $content | Should -Match '\[switch\]\$CI'
        }

        It 'Auto-installs golangci-lint when missing' {
            $content = Get-Content (Join-Path $PSScriptRoot 'lint.ps1') -Raw
            $content | Should -Match 'Install-GolangciLint'
        }
    }
}
