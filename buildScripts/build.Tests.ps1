#Requires -Version 7
<#
.SYNOPSIS
    Pester tests for buildScripts/build.ps1
#>

BeforeAll {
    # Load functions from build.ps1 without executing main logic.
    $scriptPath = Join-Path $PSScriptRoot 'build.ps1'
    $ast = [System.Management.Automation.Language.Parser]::ParseFile(
        $scriptPath, [ref]$null, [ref]$null)
    $functionDefs = $ast.FindAll(
        { param($node) $node -is [System.Management.Automation.Language.FunctionDefinitionAst] },
        $true)
    foreach ($fn in $functionDefs) {
        Invoke-Expression $fn.Extent.Text
    }
}

Describe 'build.ps1' {
    Context 'Invoke-Build function' {
        It 'Invoke-Build function is defined' {
            $cmd = Get-Command Invoke-Build -ErrorAction SilentlyContinue
            $cmd | Should -Not -BeNullOrEmpty
        }
    }

    Context 'Script structure' {
        It 'Script file exists' {
            $scriptPath = Join-Path $PSScriptRoot 'build.ps1'
            Test-Path $scriptPath | Should -Be $true
        }

        It 'Script contains ErrorActionPreference Stop' {
            $scriptPath = Join-Path $PSScriptRoot 'build.ps1'
            $content = Get-Content $scriptPath -Raw
            $content | Should -Match "\`$ErrorActionPreference\s*=\s*'Stop'"
        }

        It 'Script has comment-based help' {
            $scriptPath = Join-Path $PSScriptRoot 'build.ps1'
            $content = Get-Content $scriptPath -Raw
            $content | Should -Match '\.SYNOPSIS'
            $content | Should -Match '\.DESCRIPTION'
            $content | Should -Match '\.EXAMPLE'
            $content | Should -Match '\.NOTES'
        }
    }
}
