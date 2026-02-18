---
title: NTFS Security, Hardening and Lifecycle
author: bvlik
date: 2026-02-17
tags:
  - Security
  - NTFS
  - ACL
  - Hardening
  - Cleanup
---

# TECHNICAL DOSSIER: SECURING A HOSPITAL INFRASTRUCTURE (2/3)

> [!SUMMARY] Purpose of this document
> This document describes the implementation of the access-control policy (NTFS ACLs), the resolution of complex access scenarios, the operating-system hardening measures, and the lifecycle management of the environment (cleanup).

---

## 3. Security matrix and access control (NTFS)

Securing the hospital resources relies on the `2-Infrastructure.ps1` module. The core principle applied here is **Zero Trust (least privilege)** combined with strict directory isolation.

### 3.1. Inheritance breaking (isolation)
By default, Windows propagates the parent directory's rights to sub-folders. To make sure no standard user reaches a business folder through unintended inheritance, the script surgically breaks that chain before assigning any rights:

```powershell
# Instantiate the ACL object
$acl = Get-Acl $fullPath
# Disable inheritance (IsProtected = $true, PreserveInheritance = $true)
$acl.SetAccessRuleProtection($true, $true)
```

### 3.2. The 3 complex cases (cross-permissions)
The architecture implements at least three complex cross-permission scenarios:

> [!DANGER] Complex scenario 1: HR audit by Management
> **Need:** the `G_Direction` group must be able to read staff folders without being able to alter them.
> **Solution:** inject an explicit `ReadAndExecute` rule on the `RH` folder for Management only.
> ```powershell
> $ruleCross = New-Object System.Security.AccessControl.FileSystemAccessRule("G_Direction", "ReadAndExecute", "ContainerInherit,ObjectInherit", "None", "Allow")
> $acl.AddAccessRule($ruleCross)
> ```

> [!DANGER] Complex scenario 2: Invoice drop-box
> **Need:** the `G_Medecins` group must be able to drop accounting documents into the `Finance` folder, but must neither read the existing files nor modify the folders.
> **Solution:** create a "blind" write rule (`Write`) with no read rights.
> ```powershell
> $ruleCross = New-Object System.Security.AccessControl.FileSystemAccessRule("G_Medecins", "Write", "ContainerInherit,ObjectInherit", "None", "Allow")
> ```

> [!DANGER] Complex scenario 3: IT super-administration
> **Need:** the IT department (`G_IT`) must be able to maintain the whole tree, whoever owns the folder.
> **Solution:** apply a global, systematic `FullControl` rule for `G_IT` on every iteration of the folder-creation loop.

---

## 4. System hardening

To prevent a business user (e.g. a nurse or a doctor) from trying to bypass the restrictions via the command line, the `3-Hardening.ps1` module locks down the OS administration binaries.

### 4.1. Restricting the command interpreters
The script targets `cmd.exe` and `powershell.exe`. An explicit `Deny` rule is applied to the local `Users` group (which covers every standard account created, but does not affect the local Administrators).

```powershell
$tools = @("C:\Windows\System32\cmd.exe", "C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe")

foreach ($tool in $tools) {
    $acl = Get-Acl $tool
    # Create the execution DENY rule
    $ruleDeny = New-Object System.Security.AccessControl.FileSystemAccessRule("Users", "ReadAndExecute", "Deny")
    $acl.AddAccessRule($ruleDeny)
    Set-Acl -Path $tool -AclObject $acl
}
```
> [!NOTE] ACL precedence
> In the Windows NTFS model, a `Deny` rule always wins over an `Allow` rule. So even if the user holds native read rights, execution is blocked at the kernel level.

---

## 5. Lifecycle management and "cleanup" mode

The environment must be able to return to its initial state. The `0-Orchestrator.ps1` module handles this through an execution-parameter (`switch`) system.

### 5.1. The `-ClearAll` parameter
Running `.\0-Orchestrator.ps1 -ClearAll` triggers the de-provisioning procedure.
The script reverse-engineers the installation:
1. **Folders:** recursive removal of the tree (`Remove-Item "C:\Hopital" -Recurse -Force`).
2. **Users:** re-reads `utilisateurs.json` to target and run `Remove-LocalUser` only on the accounts created by the project (leaving vital Windows accounts untouched).
3. **Groups:** repeats the same targeting logic with `Remove-LocalGroup` through `groupes.json`.

This approach guarantees a clean, reusable lab environment for future iterations or audits.
