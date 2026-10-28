# Run only in an approved Go build environment (CI on this project).
$ErrorActionPreference = 'Stop'
$root = Split-Path $PSScriptRoot -Parent
Push-Location $root
try {
    $tool = 'github.com/josephspurrier/goversioninfo/cmd/goversioninfo@v1.7.0'
    go run $tool -64 -original-name PulseRoute.exe -o resource_windows_amd64.syso
    if ($LASTEXITCODE) { throw 'Desktop version-resource generation failed' }
    go run $tool -64 -original-name pulseroute-cli.exe -o cmd/pulseroute/resource_windows_amd64.syso
    if ($LASTEXITCODE) { throw 'CLI version-resource generation failed' }
} finally { Pop-Location }
