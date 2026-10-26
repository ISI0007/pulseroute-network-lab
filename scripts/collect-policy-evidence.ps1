# Read-only diagnostic collection. Run in an administrator-approved PowerShell.
# Does not execute PulseRoute, edit policy, change registry, or add exclusions.
$ErrorActionPreference = 'Continue'
$destination = Join-Path (Split-Path $PSScriptRoot -Parent) 'docs\evidence\policy-admin'
New-Item -ItemType Directory -Force $destination | Out-Null
$start = [DateTimeOffset]::Parse('2026-10-26T13:00:00+08:00').LocalDateTime
$end = [DateTimeOffset]::Parse('2026-10-26T14:00:00+08:00').LocalDateTime
$logs = @{
 'CodeIntegrity' = 'Microsoft-Windows-CodeIntegrity/Operational'
 'AppLocker-EXE' = 'Microsoft-Windows-AppLocker/EXE and DLL'
}
foreach ($name in $logs.Keys) {
 try {
  $events = Get-WinEvent -FilterHashtable @{ LogName=$logs[$name]; StartTime=$start; EndTime=$end } -ErrorAction Stop |
   Where-Object { $_.Message -match 'PulseRoute|pulseroute|go-build|bench.test|store.test|wire.test|network.test' -or $_.Id -eq 3099 }
  $events | Select-Object TimeCreated,Id,RecordId,ProviderName,Message | ConvertTo-Json -Depth 5 |
   Set-Content (Join-Path $destination ($name+'.json')) -Encoding UTF8
  $events | ForEach-Object { $_.ToXml() } |
   Set-Content (Join-Path $destination ($name+'-events.xml.txt')) -Encoding UTF8
 } catch { $_.Exception.Message | Set-Content (Join-Path $destination ($name+'-error.txt')) }
}
& "$env:WINDIR\System32\CiTool.exe" -lp -json | Set-Content (Join-Path $destination 'policy-inventory.json') -Encoding UTF8
Get-ItemProperty 'HKLM:\SYSTEM\CurrentControlSet\Control\CI\Policy' |
 Select-Object VerifiedAndReputablePolicyState | ConvertTo-Json |
 Set-Content (Join-Path $destination 'sac-state.json') -Encoding UTF8
Write-Output "Saved read-only evidence to $destination. Review logs for personal paths before sharing."
