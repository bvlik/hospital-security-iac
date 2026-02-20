<div align="center">

# 🏥 hospital-security-iac

**Deploy a hardened Windows 11 hospital workstation — then watch it in real time.**

Data-driven PowerShell that provisions RBAC identities, compartmentalizes folders with NTFS ACLs and
locks down the command interpreters, paired with a real-time Go SOC dashboard that proves
every change lands on the system. Defense in depth, as code.

[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
![PowerShell](https://img.shields.io/badge/PowerShell-5.1+-0A1929?style=for-the-badge&logo=powershell&logoColor=12ABDB)
![Go](https://img.shields.io/badge/Go-SOC%20TUI-0A1929?style=for-the-badge&logo=go&logoColor=12ABDB)
![Windows](https://img.shields.io/badge/Windows-11-0A1929?style=for-the-badge&logo=windows11&logoColor=12ABDB)
![Approach](https://img.shields.io/badge/Approach-Infra%20as%20Code-0070AD?style=for-the-badge)

<sub>🤝 A collaboration with <a href="https://github.com/StaiLee"><b>@StaiLee</b></a></sub>

</div>

> [!WARNING]
> **Lab use only.** The scripts create local accounts, rewrite NTFS ACLs and touch Windows
> binaries — run them in an **isolated VM**, from an **elevated PowerShell**. Never on a machine you care about.

---

## Why

Prevention without detection is half a control. This project builds a small hospital
workstation the way you'd want any sensitive endpoint built — **least privilege, isolated
data, hardened shells** — but keeps it honest by shipping the **detection** side too: a live
SOC view that turns "the ACL should block that" into "here's the 4663 event proving it did".

Everything is **data-driven**: groups, users and folders live in JSON, so adding a department
or an account is a config edit, not a code change.

## What's inside

| Layer | Tech | Job |
|-------|------|-----|
| **IAM / provisioning** | PowerShell | Local `G_*` RBAC groups + user accounts, straight from JSON |
| **ACL & isolation** | PowerShell | NTFS inheritance breaking, least privilege, 3 cross-permission scenarios |
| **Hardening** | PowerShell | Explicit `Deny` on `cmd.exe` / `powershell.exe` for standard users |
| **SOC dashboard** | Go (tview/tcell) | Live users, groups and Windows security events in a terminal UI |

## Cross-permission scenarios

The interesting part of any ACL model is where roles overlap. Three are implemented:

| # | Rule | Result |
|---|------|--------|
| 1 | `G_Direction` → **read-only** on `RH` | Management audits HR without altering it |
| 2 | `G_Medecins` → **write-only** on `Finance` | Doctors drop invoices, can't read the folder |
| 3 | `G_IT` → **FullControl** everywhere | IT maintains the tree regardless of owner |

## Layout

```
hospital-security-iac/
├── scripts/                  # Infrastructure as Code (PowerShell)
│   ├── 0-Orchestrator.ps1    #  → entry point (install / -ClearAll)
│   ├── 1-Identity.ps1        #  → RBAC groups + users (IAM)
│   ├── 2-Infrastructure.ps1  #  → folder tree + NTFS ACLs
│   ├── 3-Hardening.ps1       #  → lock down cmd / powershell
│   └── Json/                 #  → sources of truth (data-driven)
│       ├── groupes.json
│       ├── utilisateurs.json
│       └── structure.json
├── soc/                      # Real-time SOC dashboard (Go / TUI)
│   ├── main.go
│   └── go.mod
└── docs/                     # Full technical dossier (3 parts)
```

## Use it

Deploy the whole environment (elevated PowerShell):

```powershell
cd scripts
.\0-Orchestrator.ps1
```

Roll everything back — removes only what the project created:

```powershell
.\0-Orchestrator.ps1 -ClearAll
```

Run the SOC dashboard:

```powershell
cd soc
go run .
```

The dashboard exposes three panels — **👤 Users**, **👥 Groups**, **📜 Audit Trail** — refreshed
by asynchronous goroutines. Trigger a deploy in another window and watch the account/group
events (`4720` / `4726` / `4732` / `4733`) stream in live.

## Part of the blue-team toolkit

The detection side pairs naturally with the rest of the kit:

- [home-detection-lab](https://github.com/bvlik/home-detection-lab) — Sigma rules + MITRE ATT&CK scenarios + runnable detector
- [sigma-rule-pack](https://github.com/bvlik/sigma-rule-pack) — curated Sigma detections mapped to MITRE ATT&CK
- [log-triage-toolkit](https://github.com/bvlik/log-triage-toolkit) — read-only triage of auth logs (brute-force, enumeration)

## Highlights

- ✅ Data-driven RBAC, ACLs and hardening — the whole topology lives in JSON
- ✅ Idempotent, re-runnable modules with a full `-ClearAll` rollback
- ✅ Three real cross-permission scenarios (read-only / write-only / full-control)
- ✅ Command-interpreter lockdown for standard users
- ✅ Real-time Go SOC dashboard proving every change lands on the system

> Scaling this beyond a local lab (Active Directory, GPO-based hardening, SIEM
> forwarding) is discussed in the [technical dossier](docs/03-monitoring-soc.md).

---

<p align="center"><sub>Built by <b>bvlik</b> & <a href="https://github.com/StaiLee"><b>@StaiLee</b></a> — Windows security & blue-team tooling.</sub></p>
