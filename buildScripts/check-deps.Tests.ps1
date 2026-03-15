#Requires -Version 7
<#
.SYNOPSIS
    Pester tests for buildScripts/check-deps.ps1
#>

BeforeAll {
    $scriptPath = Join-Path $PSScriptRoot 'check-deps.ps1'
    $ast = [System.Management.Automation.Language.Parser]::ParseFile(
        $scriptPath, [ref]$null, [ref]$null)
    $functionDefs = $ast.FindAll(
        { param($node) $node -is [System.Management.Automation.Language.FunctionDefinitionAst] },
        $true)
    foreach ($fn in $functionDefs) {
        Invoke-Expression $fn.Extent.Text
    }
}

Describe 'check-deps.ps1' {
    Context 'Functions' {
        It 'Test-Outdated is defined' {
            Get-Command Test-Outdated -ErrorAction SilentlyContinue | Should -Not -BeNullOrEmpty
        }

        It 'Test-Vulnerabilities is defined' {
            Get-Command Test-Vulnerabilities -ErrorAction SilentlyContinue | Should -Not -BeNullOrEmpty
        }

        It 'Install-Govulncheck is defined' {
            Get-Command Install-Govulncheck -ErrorAction SilentlyContinue | Should -Not -BeNullOrEmpty
        }
    }

    Context 'Script structure' {
        It 'Script file exists' {
            Test-Path (Join-Path $PSScriptRoot 'check-deps.ps1') | Should -Be $true
        }

        It 'Has ErrorActionPreference Stop' {
            $content = Get-Content (Join-Path $PSScriptRoot 'check-deps.ps1') -Raw
            $content | Should -Match "\`$ErrorActionPreference\s*=\s*'Stop'"
        }

        It 'Has comment-based help' {
            $content = Get-Content (Join-Path $PSScriptRoot 'check-deps.ps1') -Raw
            $content | Should -Match '\.SYNOPSIS'
            $content | Should -Match '\.DESCRIPTION'
            $content | Should -Match '\.PARAMETER'
        }

        It 'Has -Outdated switch' {
            $content = Get-Content (Join-Path $PSScriptRoot 'check-deps.ps1') -Raw
            $content | Should -Match '\[switch\]\$Outdated'
        }

        It 'Has -Audit switch' {
            $content = Get-Content (Join-Path $PSScriptRoot 'check-deps.ps1') -Raw
            $content | Should -Match '\[switch\]\$Audit'
        }

        It 'Has -CI switch' {
            $content = Get-Content (Join-Path $PSScriptRoot 'check-deps.ps1') -Raw
            $content | Should -Match '\[switch\]\$CI'
        }

        It 'Auto-installs govulncheck' {
            $content = Get-Content (Join-Path $PSScriptRoot 'check-deps.ps1') -Raw
            $content | Should -Match 'Install-Govulncheck'
        }
    }
}
