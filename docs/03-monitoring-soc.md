---
title: Real-Time Monitoring (SOC) and Conclusion
author: bvlik
date: 2026-02-17
tags:
  - Security
  - SOC
  - Golang
  - Audit
  - Observability
---

# TECHNICAL DOSSIER: SECURING A HOSPITAL INFRASTRUCTURE (3/3)

> [!SUMMARY] Purpose of this document
> This document presents the advanced security layer of the project: the implementation of a real-time mini-SOC (Security Operations Center) for resource auditing, together with the final assessment of the deployment.

---

## 6. Audit and observability (real-time dashboard)

Applying NTFS security rules and hardening the system is a first layer of protection (prevention). The second, indispensable step in cybersecurity is **detection**.
To validate the effectiveness of the architecture, a real-time audit probe was developed.

### 6.1. Enabling the Windows audit policy (AuditPol)
To make the Windows kernel "talkative" about critical actions, the local audit policy was modified in depth. The system now tracks successes and failures across three major areas:

```powershell
auditpol /set /category:"Account Management" /success:enable /failure:enable
auditpol /set /category:"Object Access" /success:enable /failure:enable
auditpol /set /category:"Logon/Logoff" /success:enable /failure:enable
```
* **Account Management** — monitors privilege escalation and staff creation/deletion.
* **Object Access** — proves that the NTFS rules (ACLs) do block unauthorized access (EventID 4663 / 4624).

### 6.2. The analysis engine (Golang TUI)
Rather than forcing the administrator to dig manually through the Event Viewer, a Terminal User Interface (TUI) was developed in **Go**.

> [!INFO] Dashboard architecture
> * **Backend (Go):** uses goroutines to query the system asynchronously without blocking the interface.
> * **Frontend (tview/tcell):** renders a security matrix split into three panels: live user database, live local groups, and the security audit trail.
> * **Connector:** runs optimized PowerShell calls (`Get-WinEvent`) and parses the JSON/XML output to surgically extract the "target" of each action.

### 6.3. Tracked security events
The system filters out the noise to focus on indicators of compromise (IoC) or infrastructure changes:
* **Event 4720** — account creation (provisioning).
* **Event 4726** — account deletion (offboarding).
* **Event 4732 / 4733** — local group changes (RBAC rights change).
* **Event 4663** — file access attempt (proves the ACLs work on the medical tree).

> [!SUCCESS] Operational result
> When a deployment is launched via `0-Orchestrator.ps1`, the dashboard lights up in real time, visually categorizing each role (Doctors, IT, HR) and instantly showing the tree creation. This is the visual and technical proof that the PowerShell code really impacts the system.

---

## 7. Assessment and conclusion

The hospital infrastructure is now deployed and secured.

### 7.1. Goals achieved
1. **Privilege separation** — the RBAC model perfectly isolates the entities (Management, HR, IT, Medical).
2. **Infrastructure as Code** — the deployment is 100% automated, reproducible, and driven by data-agnostic JSON files.
3. **Defense in depth** — cross-permissions (NTFS) + executable lockdown (hardening) + traceability (Go audit).

### 7.2. Roadmap (V2)
If this project were to be ported to real production (beyond a local lab environment), the following evolutions would be required:
* **Active Directory** — migrate local accounts (`New-LocalUser`) to a centralized directory (`New-ADUser`) for multi-host management.
* **GPO (Group Policy Object)** — deploy the hardening (blocking `cmd.exe`) through domain policies rather than local ACL manipulation.
* **Log centralization** — forward security events to an external SIEM (e.g. Splunk, ELK) to avoid losing traces if the local machine is compromised.

---
*End of technical dossier.*
