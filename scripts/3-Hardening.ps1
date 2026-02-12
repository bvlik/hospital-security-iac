<#
.SYNOPSIS
    Hardening module: restricts the command interpreters.
.DESCRIPTION
    Applies an explicit Deny rule on cmd.exe and powershell.exe for the local
    Users group, preventing standard accounts from bypassing the restrictions
    through the command line.
#>
Write-Host "[*] PHASE 3 : SYSTEM HARDENING" -ForegroundColor Cyan

# Sensitive targets: command interpreters
$tools = @(
    "C:\Windows\System32\cmd.exe",
    "C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe"
)

# Note: requires prior ownership (takeown) in a real-world setup
foreach ($tool in $tools) {
    if (Test-Path $tool) {
        try {
            $acl = Get-Acl $tool
            # Explicit DENY rule (takes precedence over Allow)
            $ruleDeny = New-Object System.Security.AccessControl.FileSystemAccessRule("Utilisateurs", "ReadAndExecute", "Deny")
            $acl.AddAccessRule($ruleDeny)
            Set-Acl -Path $tool -AclObject $acl -ErrorAction Stop
            Write-Host "   [+] Lock enabled: $tool" -ForegroundColor Green
        } catch {
            Write-Host "   [!] Simulated bypass (TrustedInstaller rights required): $tool" -ForegroundColor Yellow
        }
    }
}
