<#
.SYNOPSIS
    IAM module: provisions local groups (RBAC) and user accounts.
.DESCRIPTION
    Reads groupes.json and utilisateurs.json, then creates the matching groups
    and accounts. The script is idempotent: it can be re-run without raising
    errors on objects that already exist.
#>
Write-Host "[*] PHASE 1 : IDENTITY MANAGEMENT (IAM)" -ForegroundColor Cyan

$pathG = "$PSScriptRoot\Json\groupes.json"
$pathU = "$PSScriptRoot\Json\utilisateurs.json"

# Create the local groups (idempotent)
$jsonG = Get-Content $pathG | ConvertFrom-Json
foreach ($g in $jsonG.groups) {
    if (-not (Get-LocalGroup -Name $g.name -ErrorAction SilentlyContinue)) {
        New-LocalGroup -Name $g.name -Description $g.description | Out-Null
        Write-Host "   [+] Group created: $($g.name)" -ForegroundColor Green
    }
}

# Create the users and apply the RBAC assignment
$jsonU = Get-Content $pathU | ConvertFrom-Json
foreach ($cat in $jsonU.profiles) {
    $targetGroup = "G_" + $cat.profile
    foreach ($u in $cat.users) {
        # Convert the plain-text password to a SecureString (required by New-LocalUser)
        $secPass = ConvertTo-SecureString $u.defaultPassword -AsPlainText -Force

        if (-not (Get-LocalUser -Name $u.name -ErrorAction SilentlyContinue)) {
            New-LocalUser -Name $u.name -FullName ($u.firstName + " " + $u.name) -Password $secPass -PasswordNeverExpires | Out-Null
            Write-Host "   [+] User created: $($u.name) (password: $($u.defaultPassword))" -ForegroundColor Green

            Add-LocalGroupMember -Group $targetGroup -Member $u.name
            Write-Host "       -> Added to profile: $targetGroup" -ForegroundColor Cyan
        }
    }
}
