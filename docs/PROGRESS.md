# Progress and verification
Updated 28 October 2026. Active project: D:\Project\PulseRoute\pulseroute-network-lab.

**Working development checkpoint; final native demonstration and two-device validation remain unfinished.**

## Verified CI milestone
[Run 40626295936](https://github.com/ISI0007/pulseroute-network-lab/actions/runs/40626295936), source commit 4ef6e0c52184e0817ca9b28cac6e7e00b36d7337: both jobs succeeded. Selected raw artifacts, checksums, test logs and provenance are retained under evidence/ci/4ef6e0c.

| Check | Status | Evidence and scope |
|---|---|---|
| Linux core tests with race detector | PASS | 27 test/subtest pass records, zero failures; go test -json -race ./... -count=1 |
| Windows core tests | PASS | Same 27 test/subtest outcomes; includes Go SQLite migrations/settings/history and LAN discovery over loopback |
| Windows desktop backend tests | PASS | Four tests: settings restart, bounded progress, concurrent shutdown and persisted message confirmation |
| Streamed 128 MiB file | PASS | Streaming source generation and received SHA-256/length verification; no whole-file buffer; no speed improvement claimed |
| Active cancellation and failed-history cleanup | PASS | Receiver partial/destination cleanup asserted after shutdown |
| Concurrent confirmed messages | PASS | 16 correlated acknowledgments with two client sessions; rejection and pre-send cancellation tests |
| UDP impairment handling | PASS | Actual datagram fixtures exercise delayed, duplicated, altered and missing replies plus cancellation |
| Static analysis and production builds | PASS | Linux vet; desktop vet; TypeScript/Vite; Windows CLI and Wails builds in CI |
| Real loopback experiments | PASS | 20 runs, 100 probes/run, 1024-byte payload; Windows/Linux TCP/UDP; 2000/2000 unique replies |
| Independent metric recalculation | PASS | Raw samples reproduce received count, loss, duplicates, ordering, mean RTT, jitter and echo goodput |
| Frontend six-screen smoke checks | PASS | Signed Chrome; no JS errors; no mocked Go bridge; frontend rendering only |
| Local Go execution | BLOCKED | SAC now also blocks go.exe; event 3077 record 14769; no repeat after block |
| Native Wails/WebView workflows | NOT VERIFIED | Backend tests and build do not establish native window/dialog/drop/preview behavior |
| Two-device LAN, 8 GiB, soak and controlled performance comparisons | NOT VERIFIED | Need devices and permitted native execution |
| Draft report/slides/diagrams | PASS: previous artifact checks | Generated before this CI milestone; see CI_VERIFICATION.md for the current addendum |
| GitHub synchronization | PASS | Full history preserved and published on main and codex/pulseroute-foundation |
| License | PASS | Owner approved MIT; root LICENSE published |
| Free signing sponsorship | PENDING | Policy/application drafts prepared; no acceptance or signature |
| Academic identity | PENDING | Owner's details not supplied |

## Current features
Node settings, known peers and recent benchmark results restore from SQLite. Shutdown cancels tracked operations and rejects new database work. Transfer state and preview reads are bounded. Handshakes cancel promptly. UDP replies must match a per-experiment cookie and payload. Desktop Send waits for server acceptance and displays the socket-write count; this is not human-read confirmation. Lost acknowledgment outcomes remain unknown and are not automatically retried.

## Evidence and limits
Raw logs include test names and outcomes. Experiment CSV, plots, reproduction command and interpretation limits are in experiments/4ef6e0c. Hosted loopback is not physical link capacity or a controlled Windows-versus-Linux comparison. Existing report/PPT drafts predate these results and remain drafts. Hub file relay/download, resume/crash reconciliation, encryption and authentication are not implemented.

## Historical failures
PowerShell npm.ps1 authorization and sandbox TEMP failures were resolved by the ordinary npm.cmd entry point and authorized D: temporary directories. Earlier browser and event-log sandbox restrictions were resolved through approved read-only/tool execution. The new compiler SAC block was not bypassed; CI supplied an independent approved runtime. No security controls were changed.

## Free signing preparation
The owner approved MIT. Both Windows executables now carry consistent product/version metadata, verified in successful run 40628906942 (e6d1f83). Signing policy, privacy/removal guidance, dependency notice inventory and an unsubmitted application draft are present. Private keys, personal enrollment facts and contact details are not published. A provider application and an official release still require owner review/approval.
