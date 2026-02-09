<#
.SYNOPSIS
    Filesystem module: builds the folder tree and applies NTFS ACLs.
.DESCRIPTION
    Creates the business folders defined in structure.json, breaks NTFS
    inheritance to guarantee isolation, then applies the base access rules
    (least privilege) along with three cross-permission scenarios.
#>
Write-Host "[*] PHASE 2 : INFRASTRUCTURE & ACL DEPLOYMENT" -ForegroundColor Cyan

$pathS = "$PSScriptRoot\Json\structure.json"
$jsonS = Get-Content $pathS | ConvertFrom-Json

foreach ($folder in $jsonS.folders) {
    # Create the physical folder
    if (-not (Test-Path $folder.path)) {
        New-Item -ItemType Directory -Path $folder.path | Out-Null
    }

    $acl = Get-Acl $folder.path

    # Isolation: break NTFS inheritance (least-privilege principle)
    $acl.SetAccessRuleProtection($true, $false)

    # Base rule (business owner)
    $ruleBase = New-Object System.Security.AccessControl.FileSystemAccessRule($folder.group, $folder.rights, "ContainerInherit,ObjectInherit", "None", "Allow")
    $acl.AddAccessRule($ruleBase)

    # Global rule (IT keeps FullControl everywhere)
    $ruleIT = New-Object System.Security.AccessControl.FileSystemAccessRule("G_IT", "FullControl", "ContainerInherit,ObjectInherit", "None", "Allow")
    $acl.AddAccessRule($ruleIT)

    # --- COMPLEX CROSS-PERMISSION SCENARIOS ---

    # Case 1: Management audits HR (read-only)
    if ($folder.path -match "RH") {
        $ruleCross1 = New-Object System.Security.AccessControl.FileSystemAccessRule("G_Direction", "ReadAndExecute", "ContainerInherit,ObjectInherit", "None", "Allow")
        $acl.AddAccessRule($ruleCross1)
    }

    # Case 2: Doctors drop invoices in Finance (write-only / blind write)
    if ($folder.path -match "Finance") {
        $ruleCross2 = New-Object System.Security.AccessControl.FileSystemAccessRule("G_Medecins", "Write", "ContainerInherit,ObjectInherit", "None", "Allow")
        $acl.AddAccessRule($ruleCross2)
    }

    Set-Acl -Path $folder.path -AclObject $acl
    Write-Host "   [+] Secured folder: $($folder.path)" -ForegroundColor Green
}
