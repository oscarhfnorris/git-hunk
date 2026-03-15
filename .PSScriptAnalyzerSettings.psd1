@{
    # Rules to include (empty means all rules).
    IncludeRules = @()

    # Rules to exclude.
    ExcludeRules = @(
        'PSAvoidUsingWriteHost'   # We intentionally use Write-Host for coloured output.
    )

    # Severity levels to report.
    Severity = @('Error', 'Warning', 'Information')
}
