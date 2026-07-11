[CmdletBinding()]
param(
    [Parameter(Position = 0)]
    [string]$Remote = "origin",

    [Parameter(Position = 1)]
    [string]$Version
)

$ErrorActionPreference = "Stop"

function Invoke-Git {
    param(
        [Parameter(Mandatory)]
        [string[]]$Arguments
    )

    & git @Arguments
    if ($LASTEXITCODE -ne 0) {
        throw "git $($Arguments -join ' ') failed with exit code $LASTEXITCODE"
    }
}

try {
    Invoke-Git -Arguments @("rev-parse", "--is-inside-work-tree")

    if ([string]::IsNullOrWhiteSpace($Remote)) {
        $Remote = "origin"
    }

    & git remote get-url $Remote | Out-Null
    if ($LASTEXITCODE -ne 0) {
        throw "Remote repository '$Remote' does not exist."
    }

    $Branch = (& git branch --show-current).Trim()
    if ($LASTEXITCODE -ne 0 -or [string]::IsNullOrWhiteSpace($Branch)) {
        throw "Unable to determine the current branch. Check out a branch before running this script."
    }

    & git check-ref-format --branch $Branch | Out-Null
    if ($LASTEXITCODE -ne 0) {
        throw "Invalid branch name: $Branch"
    }

    Write-Host "Fetching $Remote/$Branch..."
    Invoke-Git -Arguments @("fetch", $Remote, $Branch)

    Write-Host "Rebasing HEAD onto $Remote/$Branch..."
    & git rebase "$Remote/$Branch"
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Rebase stopped because of a conflict or another error. Resolve it and run 'git rebase --continue', or abort with 'git rebase --abort'. Nothing was pushed."
        exit $LASTEXITCODE
    }

    Write-Host "Pushing HEAD to $Remote/$Branch and setting upstream..."
    Invoke-Git -Arguments @(
        "push",
        "--set-upstream",
        "--force-with-lease=refs/heads/$Branch",
        $Remote,
        "HEAD:refs/heads/$Branch"
    )

    if ([string]::IsNullOrWhiteSpace($Version)) {
        Write-Host "Push completed. No version was supplied, so no tag was created."
        exit 0
    }

    & git check-ref-format "refs/tags/$Version"
    if ($LASTEXITCODE -ne 0) {
        throw "Invalid tag name: $Version"
    }

    Write-Host "Force-setting tag $Version to HEAD..."
    Invoke-Git -Arguments @("tag", "--force", $Version, "HEAD")

    Write-Host "Force-pushing tag $Version to $Remote..."
    Invoke-Git -Arguments @("push", "--force", $Remote, "refs/tags/${Version}:refs/tags/${Version}")

    Write-Host "Push and tag update completed successfully."
}
catch {
    Write-Error $_
    exit 1
}
