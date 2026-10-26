$ErrorActionPreference = 'Stop'
Set-Location (Split-Path $PSScriptRoot -Parent)
New-Item -ItemType Directory -Force docs/evidence | Out-Null
go test ./... -count=1 -v 2>&1 | Tee-Object docs/evidence/go-tests.txt
if ($LASTEXITCODE -ne 0) { throw 'Go tests failed or were blocked. Inspect go-tests.txt.' }
go vet ./...
if ($LASTEXITCODE -ne 0) { throw 'Static analysis failed' }
Push-Location frontend
try { npm.cmd ci; if ($LASTEXITCODE -ne 0) { throw 'npm ci failed' }; npm.cmd run build; if ($LASTEXITCODE -ne 0) { throw 'Frontend build failed' } } finally { Pop-Location }
go vet -tags desktop .
if ($LASTEXITCODE -ne 0) { throw 'Desktop static analysis failed' }
go build -tags 'desktop,production' -ldflags '-H windowsgui' -o build/bin/PulseRoute.exe .
if ($LASTEXITCODE -ne 0) { throw 'Desktop build failed' }
go build -o build/bin/pulseroute-cli.exe ./cmd/pulseroute
if ($LASTEXITCODE -ne 0) { throw 'CLI build failed' }
