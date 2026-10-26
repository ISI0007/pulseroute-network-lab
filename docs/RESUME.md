# PulseRoute restart checkpoint
User explicitly requested a stop after Windows policy diagnosis, due to their usage limit. Resume only when asked.

## Active locations
- Project: D:\Project\PulseRoute\pulseroute-network-lab
- Go: D:\Project\PulseRoute\work\go\bin
- GOPATH: D:\Project\PulseRoute\work\modules
- GOCACHE: D:\Project\PulseRoute\work\gocache
- TEMP/TMP: D:\Project\PulseRoute\work\tmp
- Browser test dependency: D:\Project\PulseRoute\work\browser\node_modules\playwright
- Python document libraries: D:\Project\PulseRoute\work\python-libs (CPython 3.13); use D:\tools\venv\Scripts\python.exe with PYTHONPATH set to that directory.
Some relocation remnants remain in the original C: workspace; the active project and binaries are on D:. Do not delete unrelated files.

## Work completed
Go framing/chat/file transfer, CLI, UDP echo metrics, LAN discovery, SQLite migration/store, Wails bridge, React/TypeScript/Tailwind UI and compileable Windows binaries. Early core tests passed, then Windows blocked later executable launches. Current source needs a full regression run. Initial local commit: 54f499c. A later checkpoint commit preserves the rest; no GitHub push succeeded.

## Important blockers and permissions
Smart App Control registry state is 1 (enforcement). Policy attribution is probable, not proven by event records, because event logs, CIM and CiTool inventory access were denied. See WINDOWS_APPLICATION_CONTROL.md. User explicitly prohibits disabling/bypassing controls and unnecessary repeats of blocked commands. Never infer that conversational approval changes OS policy.
GitHub connector user is ISI0007; exact target ISI0007/pulseroute-network-lab remains 404 even after user reported creating/granting access. Connector cannot create repositories. Local gh authentication invalid; do not handle credentials without approval.
License requested MIT/Apache/unlicensed; no answer. Do not invent a choice.
D:\Project filesystem permission was granted for the current turn; on a later turn verify access and request the specific project path only if necessary.

## Code review focus before continuing
- Reconnect cancellation and shutdown concurrency need actual race tests.
- Persistence schema is event-oriented; messages/file metadata live in JSON payloads rather than separate normalized tables.
- UI bridge is compile-checked but untested at runtime; no actual native screenshots.
- UDP uses sequential echo with late-response observations, not saturated throughput or one-way loss/jitter.
- Received files use random prefixes and SHA-256; no resume or crash reconciliation.
- Snapshot history is bounded to latest 200 events; duplicate server/client observations are deduplicated in the conversation view.
- Settings currently persist theme; endpoints/name are session-only. Benchmark sessions persist, but historical results are not reloaded into the UI yet.
- Direct transfers target receiver listeners. A hub can receive a file, but hub-to-another-client file relay/download is not implemented.
- Native local file selection is required before SendFile; inline media preview limited to 16 MiB.
- No encryption/authentication, internet P2P, live video, or congestion-control implementation.

## Next steps in order
1. Finish read-only policy attribution using administrator-collected event XML.
2. Obtain an approved execution/signing or CI route, then run go test ./... and race checks. Do not repeat blocked commands without a relevant change.
3. Verify GitHub access before reconciling remote initial README; preserve remote work and use a codex branch. Never force-push.
4. Test SQLite, discovery, reconnect, UI queue/cancel, previews, persistence and two-device transfer.
5. Implement missing hub file routing, stored settings/history reload, stronger recovery as needed.
6. Capture genuine screenshots and reproducible raw benchmarks.
7. Complete report DOCX/PDF, ten diagrams, PowerPoint with notes, live demo and packaging.
8. Update status/limitations honestly and commit verified increments.

## Commands (after prerequisites are satisfied)
Set PATH, GOPATH, GOCACHE and TEMP to the D: paths above. Go module requires 1.26.0+ (tested compiler 1.27.3).
- go test ./... -count=1 -v
- go vet ./...
- cd frontend; npm.cmd ci; npm.cmd run build
- go vet -tags desktop .
- go build -tags desktop,production -ldflags "-H windowsgui" -o build/bin/PulseRoute.exe .
- go build -o build/bin/pulseroute-cli.exe ./cmd/pulseroute
- CLI: -mode serve; -mode chat; -mode send -file PATH; -mode bench -protocol tcp/udp; -mode selftest
- scripts/verify.ps1 stops on failure; scripts/inspect-ui.cjs checks only the built frontend with no mocked backend.
