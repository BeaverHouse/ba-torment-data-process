# Script for cleaning up Git branches
# This is PowerShell script, see clean-git-branch.sh for the Bash version.
# Written by Austin Lee.

Write-Host "Fetching and pruning remote branches..." -ForegroundColor Green
git remote update origin --prune
git config --global advice.forceDeleteBranch false

# Get local and remote branch information
$localBranches = git branch --format="%(refname:short)" | Out-String
$remoteBranches = git branch -r | Out-String

# Check each local branch
$localBranches.Split("`n") | ForEach-Object {
    $branch = $_.Trim()
    if ($branch) {
        $remoteBranch = "origin/$branch"
        if (-not ($remoteBranches -match [regex]::Escape($remoteBranch))) {
            git branch -D $branch
            Write-Host "Deleted local branch: $branch" -ForegroundColor Yellow
        }
        else {
            Write-Host "Local branch $branch exists in remote" -ForegroundColor Cyan
        }
    }
}