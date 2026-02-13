<#
.SYNOPSIS
    Single entry point for deploying the hospital infrastructure.
.DESCRIPTION
    Orchestrates the IAM, Infrastructure/ACL and Hardening modules.
    The -ClearAll switch rolls the environment back to its initial state.
.EXAMPLE
    .\0-Orchestrator.ps1
    Deploys the whole infrastructure.
.EXAMPLE
    .\0-Orchestrator.ps1 -ClearAll
    Removes the folders, users and groups created by the project.
#>
param (
    [switch]$ClearAll
)

$ScriptPath = $PSScriptRoot

if ($ClearAll) {
    Write-Host "[!] CLEANUP MODE ENABLED" -ForegroundColor Red

    # 1. Remove the folders
    if (Test-Path "C:\Hopital") { Remove-Item "C:\Hopital" -Recurse -Force }

    # 2. Remove only the users created by the project
    $jsonU = Get-Content "$ScriptPath\Json\utilisateurs.json" | ConvertFrom-Json
    foreach ($cat in $jsonU.profiles) {
        foreach ($u in $cat.users) {
            Remove-LocalUser -Name $u.name -ErrorAction SilentlyContinue
        }
    }

    # 3. Remove only the groups created by the project
    $jsonG = Get-Content "$ScriptPath\Json\groupes.json" | ConvertFrom-Json
    foreach ($g in $jsonG.groups) {
        Remove-LocalGroup -Name $g.name -ErrorAction SilentlyContinue
    }

    Write-Host "[+] Cleanup complete." -ForegroundColor Green
    exit
}

Write-Host "DEPLOYMENT STARTED" -ForegroundColor Magenta
& "$ScriptPath\1-Identity.ps1"
& "$ScriptPath\2-Infrastructure.ps1"
& "$ScriptPath\3-Hardening.ps1"
Write-Host "DEPLOYMENT COMPLETE" -ForegroundColor Magenta
