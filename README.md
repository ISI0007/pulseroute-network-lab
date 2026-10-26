# PulseRoute
### High-Performance Multimedia Transfer and Network Performance Laboratory

A Windows-focused Computer Networks course project, written in Go with a planned React/TypeScript/Wails desktop shell.

**Development build — consult [verification](docs/PROGRESS.md) before treating a feature as complete.**
No measured performance claim is made until reproducible evidence is recorded.

## Structure
- `internal/wire`: bounded length-prefixed frames.
- `internal/network`: concurrent TCP chat and streamed file transfer.
- `internal/bench`: TCP/UDP experiments and raw measurements.
- `internal/store`: SQLite persistence and migrations.
- `frontend`: desktop interface.
- `docs`: specifications, architecture, evidence, academic materials.

## Build and test
Install Go 1.24 or newer, Node.js, and the Wails v2 prerequisites.
From the project root:
```powershell
go test ./...
go vet ./...
go test -race ./...
```
The race detector requires a supported C toolchain on Windows. Run the provided CI workflow on Linux if unavailable locally.
Desktop build commands will be recorded after toolchain verification.

## Scope and trust boundary
This laboratory is intended for trusted local networks. TCP provides an ordered byte stream; application framing and SHA-256 validate complete file transfers. UDP experiments measure application-observed behavior, not packet captures. Video files are transferred as binary files; live video calling is outside the implemented scope. Internet NAT traversal and custom TCP congestion-control algorithms are not claimed.

## License
Pending the project owner's explicit choice. No open-source license has been granted by this repository yet.
