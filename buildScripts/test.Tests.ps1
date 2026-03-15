#Requires -Version 7
<#
.SYNOPSIS
    Pester tests for buildScripts/test.ps1
#>

BeforeAll {
    $scriptPath = Join-Path $PSScriptRoot 'test.ps1'
    $ast = [System.Management.Automation.Language.Parser]::ParseFile(
        $scriptPath, [ref]$null, [ref]$null)
    $functionDefs = $ast.FindAll(
        { param($node) $node -is [System.Management.Automation.Language.FunctionDefinitionAst] },
        $true)
    foreach ($fn in $functionDefs) {
        Invoke-Expression $fn.Extent.Text
    }
}

Describe 'test.ps1' {
    Context 'Invoke-Tests function' {
        It 'Invoke-Tests function is defined' {
            $cmd = Get-Command Invoke-Tests -ErrorAction SilentlyContinue
            $cmd | Should -Not -BeNullOrEmpty
        }
    }

    Context 'Script structure' {
        It 'Script file exists' {
            Test-Path (Join-Path $PSScriptRoot 'test.ps1') | Should -Be $true
        }

        It 'Script contains ErrorActionPreference Stop' {
            $content = Get-Content (Join-Path $PSScriptRoot 'test.ps1') -Raw
            $content | Should -Match "\`$ErrorActionPreference\s*=\s*'Stop'"
        }

        It 'Script has comment-based help' {
            $content = Get-Content (Join-Path $PSScriptRoot 'test.ps1') -Raw
            $content | Should -Match '\.SYNOPSIS'
            $content | Should -Match '\.DESCRIPTION'
            $content | Should -Match '\.PARAMETER'
            $content | Should -Match '\.EXAMPLE'
        }

        It 'Script has -Coverage switch' {
            $content = Get-Content (Join-Path $PSScriptRoot 'test.ps1') -Raw
            $content | Should -Match '\[switch\]\$Coverage'
        }

        It 'Script has -CI switch' {
            $content = Get-Content (Join-Path $PSScriptRoot 'test.ps1') -Raw
            $content | Should -Match '\[switch\]\$CI'
        }
    }
}
