# PulseRoute
### High-Performance Multimedia Transfer and Network Performance Laboratory

Go networking engines with a React/TypeScript/Tailwind interface and Wails desktop shell, developed for a Computer Networks course.

**Unfinished development checkpoint.** Start with [progress](docs/PROGRESS.md), [restart notes](docs/RESUME.md), and [Windows policy diagnosis](docs/WINDOWS_APPLICATION_CONTROL.md). Successful compilation does not establish that desktop workflows work.

## Implemented source
- Bounded length-prefixed TCP framing, concurrent chat, acknowledgments and client connections.
- Streaming file transfer, SHA-256 verification, temporary-file cleanup, cancellation and progress.
- TCP/UDP echo experiments, sequence-aware observations and raw-data structures.
- LAN discovery and direct TCP file targeting.
- SQLite schema/migration, settings, peer/event history and benchmark records.
- Desktop overview, messages, transfer queue, media preview, lab, history and settings.

Earlier tests passed for the core protocol, eight clients, file identity up to 16 MiB, interruption cleanup, and TCP/UDP loopback replies. Current core tests now pass on Linux (with race detection) and Windows in GitHub CI, including a 128 MiB streamed integrity test and four Windows desktop backend tests. Local execution remains blocked by Smart App Control. Twenty hosted loopback runs and raw samples are retained; native GUI and two-device LAN validation remain pending. See [CI verification](docs/CI_VERIFICATION.md).

## Build
Go 1.26+ (compiler used: 1.27.3), Node.js 24, WebView2, Windows amd64.
```powershell
go test ./...
go vet ./...
cd frontend
npm.cmd ci
npm.cmd run build
cd ..
go vet -tags desktop .
go build -tags desktop,production -ldflags "-H windowsgui" -o build/bin/PulseRoute.exe .
go build -o build/bin/pulseroute-cli.exe ./cmd/pulseroute
```
Both executables were compiled locally. Builds are unsigned. Use the approved signing/policy process before execution on protected systems; do not disable security controls.

## CLI examples
```powershell
.\build\bin\pulseroute-cli.exe -mode serve -addr 127.0.0.1:42191 -udp 127.0.0.1:42192
.\build\bin\pulseroute-cli.exe -mode chat -addr 127.0.0.1:42191 -name Alice
.\build\bin\pulseroute-cli.exe -mode send -addr 127.0.0.1:42191 -file C:\test\sample.bin
.\build\bin\pulseroute-cli.exe -mode bench -protocol udp -udp 127.0.0.1:42192
```

## Boundaries
Trusted local networks only; no encryption or peer authentication. Video files are transferred, not live video calls. No NAT traversal, custom TCP congestion control, transfer resume or server-mediated file forwarding to another client. Echo goodput is not link capacity. SQLite runtime and desktop bindings still need verification.

## Project map
`internal/wire`, `internal/network`, `internal/bench`, `internal/store`; `cmd/pulseroute`; `app.go` and `main.go` (desktop build tag); `frontend`; `docs`; `scripts`.
Versioned SQLite schema: internal/store/migrations/001_initial.sql.
Repository access was restored on 28 October 2026 and the full checkpoint was pushed. The verification workflow runs on pushes; consult GitHub Actions for its actual result.

License is pending the owner's explicit choice. No LICENSE has been added.

## Draft academic deliverables
[Technical report PDF](deliverables/PulseRoute-Technical-Report-DRAFT.pdf), [editable DOCX](deliverables/PulseRoute-Technical-Report-DRAFT.docx), [presentation with notes](deliverables/PulseRoute-Presentation-DRAFT.pptx), and [ten diagrams](docs/diagrams). These are progress drafts, not a completed submission. See [verification status](docs/PROGRESS.md) for the exact evidence boundary.

## Recorded measurements
[Recalculated loopback results](docs/experiments/4ef6e0c/RESULTS.md) include raw sample provenance, CSV and measured plots. These are sequential application echo results, not physical link capacity.

## Signed distribution
[Signing and portable packaging](docs/DISTRIBUTION.md) is prepared. Trusted signing is pending an eligible verified publisher identity; existing build artifacts remain unsigned. Zehri Developers is the product brand, not yet a verified certificate publisher.
