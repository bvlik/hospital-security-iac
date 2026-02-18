---
title: Architecture and Identity Management (IAM)
author: bvlik
date: 2026-02-17
tags:
  - Security
  - Windows11
  - PowerShell
  - IAM
  - ACL
---

# TECHNICAL DOSSIER: SECURING A HOSPITAL INFRASTRUCTURE (1/3)

> [!SUMMARY] Purpose of this document
> This document describes the software architecture and the identity management (IAM) policy put in place to secure the Windows 11 environment of the hospital.

---

## 1. Deployment architecture (Infrastructure as Code)

To meet the requirement for automation and maintainability, the deployment does not rely on a monolithic script but on a modular architecture driven by external sources of truth (JSON files).

### 1.1. Data model (JSON files)
The hospital topology is fully extracted from the business code. Three configuration files drive the behaviour of the scripts:

1. **`groupes.json`** — defines the RBAC (Role-Based Access Control) matrix through local groups prefixed with `G_` (e.g. `G_Medecins`, `G_Direction`).
2. **`utilisateurs.json`** — lists the staff, their assignment, and their initial provisioning password.
3. **`structure.json`** — maps the folder tree and defines the primary access rights (FullControl, Modify, etc.).

### 1.2. Modular PowerShell architecture
To keep the code readable, auditable and aligned with development standards, the actions are split into modules driven by an orchestrator:

* `0-Orchestrator.ps1` — single entry point. Handles the execution parameters (install vs. cleanup) and calls the sub-modules.
* `1-Identity.ps1` — IAM module (creation of local users and groups).
* `2-Infrastructure.ps1` — filesystem module (folder tree creation and complex NTFS/ACL rules).
* `3-Hardening.ps1` — hardening module (restriction of critical executables).

---

## 2. Identity and Access Management (IAM)

The provisioning of local accounts is handled by the `1-Identity.ps1` module. It strictly applies the security guidelines and the required technical specifics.

### 2.1. Group provisioning (RBAC)
The script reads `groupes.json` and checks that the group does not already exist through a conditional block (`-not (Get-LocalGroup ...)`). This makes the script **idempotent**: it can be re-run multiple times without raising interrupting errors.

### 2.2. User creation and the password subtlety
Provisioning passwords involves a major technical constraint: **ingesting a plain-text password from a data source**.

> [!WARNING] Technical justification
> The `New-LocalUser` cmdlet requires a `SecureString` object for the `-Password` parameter. However, the source data (`defaultMotDePasse` in the JSON) is a standard string.

To satisfy this constraint, the script forces the conversion of the plain string into a secure object before creating the account, using the `-AsPlainText -Force` options:

```powershell
# Read the plain-text password from the JSON configuration file
$plainPassword = $u.defaultPassword

# Mandatory conversion to satisfy New-LocalUser
$secPass = ConvertTo-SecureString $plainPassword -AsPlainText -Force

# Create the user with a non-expiring password (lab/demo environment)
New-LocalUser -Name $u.name -FullName ($u.firstName + " " + $u.name) -Password $secPass -PasswordNeverExpires
```

> [!SUCCESS] Output confirmation
> During execution, the script explicitly prints the initial password to the console (e.g. `[+] User created: medecin1 (password: P@ssword1)`), letting the administrator confirm that the configuration file was correctly applied at first login.

### 2.3. Dynamic assignment
Users are automatically linked to their business group. The script concatenates the `G_` prefix with the profile read from the JSON to enforce strict consistency between identity and rights (e.g. the "Medecins" profile is automatically mapped to the `G_Medecins` group).
