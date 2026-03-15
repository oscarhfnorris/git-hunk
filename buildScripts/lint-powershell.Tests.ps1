#Requires -Version 7
<#
.SYNOPSIS
    Pester tests for buildScripts/lint-powershell.ps1
#>

BeforeAll {
    $scriptPath = Join-Path $PSScriptRoot 'lint-powershell.ps1'
    $ast = [System.Management.Automation.Language.Parser]::ParseFile(
        $scriptPath, [ref]$null, [ref]$null)
    $functionDefs = $ast.FindAll(
        { param($node) $node -is [System.Management.Automation.Language.FunctionDefinitionAst] },
        $true)
    foreach ($fn in $functionDefs) {
        Invoke-Expression $fn.Extent.Text
    }
}

Describe 'lint-powershell.ps1' {
    Context 'Functions' {
        It 'Invoke-PSLint is defined' {
            Get-Command Invoke-PSLint -ErrorAction SilentlyContinue | Should -Not -BeNullOrEmpty
        }

        It 'Install-PSScriptAnalyzer is defined' {
            Get-Command Install-PSScriptAnalyzer -ErrorAction SilentlyContinue | Should -Not -BeNullOrEmpty
        }
    }

    Context 'Script structure' {
        It 'Script file exists' {
            Test-Path (Join-Path $PSScriptRoot 'lint-powershell.ps1') | Should -Be $true
        }

        It 'Has ErrorActionPreference Stop' {
            $content = Get-Content (Join-Path $PSScriptRoot 'lint-powershell.ps1') -Raw
            $content | Should -Match "\`$ErrorActionPreference\s*=\s*'Stop'"
        }

        It 'Has comment-based help with .NOTES' {
            $content = Get-Content (Join-Path $PSScriptRoot 'lint-powershell.ps1') -Raw
            $content | Should -Match '\.NOTES'
        }

        It 'Reads settings from .PSScriptAnalyzerSettings.psd1' {
            $content = Get-Content (Join-Path $PSScriptRoot 'lint-powershell.ps1') -Raw
            $content | Should -Match '\.PSScriptAnalyzerSettings\.psd1'
        }

        It 'Has -Fix switch' {
            $content = Get-Content (Join-Path $PSScriptRoot 'lint-powershell.ps1') -Raw
            $content | Should -Match '\[switch\]\$Fix'
        }

        It 'Has -CI switch' {
            $content = Get-Content (Join-Path $PSScriptRoot 'lint-powershell.ps1') -Raw
            $content | Should -Match '\[switch\]\$CI'
        }

        It 'Auto-installs PSScriptAnalyzer' {
            $content = Get-Content (Join-Path $PSScriptRoot 'lint-powershell.ps1') -Raw
            $content | Should -Match 'Install-PSScriptAnalyzer'
        }
    }

    Context 'Settings file' {
        It '.PSScriptAnalyzerSettings.psd1 exists at project root' {
            $settingsPath = Join-Path $PSScriptRoot '..' '.PSScriptAnalyzerSettings.psd1'
            Test-Path $settingsPath | Should -Be $true
        }
    }
}
