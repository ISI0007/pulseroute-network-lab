[CmdletBinding()]
param(
    [Parameter(Mandatory=$true)][ValidatePattern('^[0-9a-fA-F]{40}$')][string]$SourceCommit,
    [string]$InputDirectory = (Join-Path $PSScriptRoot '../build/bin'),
    [string]$OutputDirectory = (Join-Path $PSScriptRoot '../build/distribution'),
    [string]$CertificateThumbprint,
    [string]$SignToolPath,
    [string]$TimestampServer = 'http://timestamp.digicert.com',
    [switch]$VerifyOnly
)
# Uses an existing approved signing identity; never exports keys or edits trust.
$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest
$names = @('PulseRoute.exe', 'pulseroute-cli.exe')
$inputRoot = (Resolve-Path -LiteralPath $InputDirectory).Path
foreach ($name in $names) {
    if (-not (Test-Path -LiteralPath (Join-Path $inputRoot $name) -PathType Leaf)) { throw "Missing executable: $name" }
}
if (-not $VerifyOnly) {
    $thumbprint = ($CertificateThumbprint -replace ' ', '')
    if ($thumbprint -notmatch '^[0-9a-fA-F]{40}$') { throw 'Provide the thumbprint of an approved code-signing certificate in CurrentUser/My, or use VerifyOnly for externally signed binaries.' }
    $certificate = Get-Item -LiteralPath "Cert:\CurrentUser\My\$thumbprint"
    if (-not $certificate.HasPrivateKey) { throw 'The selected certificate has no accessible private key.' }
    if ($certificate.NotBefore -gt (Get-Date) -or $certificate.NotAfter -le (Get-Date)) { throw 'Signing certificate is not currently valid.' }
    if ($certificate.EnhancedKeyUsageList.ObjectId -notcontains '1.3.6.1.5.5.7.3.3') { throw 'Certificate is not valid for code signing.' }
    if ($certificate.Subject -eq $certificate.Issuer) { throw 'Self-signed certificates are not accepted for this distribution package.' }
} else {
    foreach ($name in $names) {
        $signature = Get-AuthenticodeSignature -LiteralPath (Join-Path $inputRoot $name)
        if ($signature.Status -ne 'Valid' -or $null -eq $signature.TimeStamperCertificate -or $signature.SignerCertificate.Subject -eq $signature.SignerCertificate.Issuer) { throw "$name requires a valid timestamped Authenticode signature; no package was produced." }
    }
}
if (-not $SignToolPath) {
    $command = Get-Command signtool.exe -ErrorAction SilentlyContinue
    if ($command) { $SignToolPath = $command.Source }
}
if (-not $SignToolPath -or -not (Test-Path -LiteralPath $SignToolPath -PathType Leaf)) { throw 'Specify SignToolPath pointing to signtool.exe from the official Windows SDK.' }
$stage = Join-Path $OutputDirectory ('staging-' + [guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $stage -Force | Out-Null
$records = @()
foreach ($name in $names) {
    $target = Join-Path $stage $name
    Copy-Item -LiteralPath (Join-Path $inputRoot $name) -Destination $target
    if (-not $VerifyOnly) {
        & $SignToolPath sign /s My /sha1 $thumbprint /fd SHA256 /tr $TimestampServer /td SHA256 /d 'PulseRoute - Zehri Developers' $target
        if ($LASTEXITCODE -ne 0) { throw "Signing failed for $name. No ZIP will be produced." }
    }
    & $SignToolPath verify /pa /all /tw /v $target
    if ($LASTEXITCODE -ne 0) { throw "Signature verification failed for $name. No ZIP will be produced." }
    $signature = Get-AuthenticodeSignature -LiteralPath $target
    if ($signature.Status -ne 'Valid' -or $null -eq $signature.TimeStamperCertificate -or $signature.SignerCertificate.Subject -eq $signature.SignerCertificate.Issuer) { throw "$name lacks valid timestamped Authenticode evidence." }
    $records += [ordered]@{ file=$name; sha256=(Get-FileHash -LiteralPath $target -Algorithm SHA256).Hash; signatureStatus=[string]$signature.Status; publisher=$signature.SignerCertificate.Subject; certificateThumbprint=$signature.SignerCertificate.Thumbprint; timestampAuthority=$signature.TimeStamperCertificate.Subject }
}
$guide = Join-Path $PSScriptRoot '../docs/DISTRIBUTION.md'
Copy-Item -LiteralPath $guide -Destination (Join-Path $stage 'READ-ME.md')
[ordered]@{sourceCommit=$SourceCommit; createdAtUTC=[DateTime]::UtcNow.ToString('o'); format='portable Windows x64 ZIP; not a setup installer'; files=$records; validation='Authenticode policy and timestamp verification on packaging host; target Smart App Control/native/LAN tests remain required'} | ConvertTo-Json -Depth 6 | Set-Content -Encoding UTF8 (Join-Path $stage 'signature-manifest.json')
$archive = Join-Path $OutputDirectory ('PulseRoute-windows-x64-' + $SourceCommit.Substring(0,8) + '-' + [guid]::NewGuid().ToString('N').Substring(0,8) + '.zip')
Compress-Archive -LiteralPath (Join-Path $stage 'PulseRoute.exe'),(Join-Path $stage 'pulseroute-cli.exe'),(Join-Path $stage 'READ-ME.md'),(Join-Path $stage 'signature-manifest.json') -DestinationPath $archive
Get-FileHash -LiteralPath $archive -Algorithm SHA256 | Select-Object Path,Hash,Algorithm
