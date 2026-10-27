# Progress and verification
Updated 28 October 2026 after the user's request to continue. Active project: D:\Project\PulseRoute\pulseroute-network-lab.

**Unfinished development checkpoint, not a final submission.** PASS means the named check succeeded; FAIL means an attempted check failed; BLOCKED means a restriction prevented validation; NOT VERIFIED means no relevant successful runtime test.

| Check | Status | Scope |
|---|---|---|
| TCP framing, eight clients, file integrity and cleanup | PASS at earlier revision | Historical core tests; not regression evidence for current source |
| TCP/UDP loopback | PASS at earlier revision | 20/20 replies each; no raw performance series retained |
| Current Go runtime suite and CLI self-test | BLOCKED | Smart App Control event 3077; not rerun |
| Go static checks | PASS | go vet ./... and go vet -tags desktop . |
| Internal and desktop test compilation | PASS | go test -c; binaries were NOT executed |
| CLI and desktop production compilation | PASS | Final builds include current frontend; unsigned |
| TypeScript/Vite build | PASS | npm.cmd run build; 1,578 modules |
| SQL schema validation | PASS | Python SQLite actual migration: rollback, foreign keys, reopen, integrity; not Go driver validation |
| Frontend six-screen smoke checks | PASS | Installed signed Chrome; no JS errors; bridge unavailable, no mocked application data |
| Desktop bridge, settings restart, discovery, reconnect | NOT VERIFIED | Implementation and tests compile; native runtime blocked |
| Race detector | BLOCKED | CGO/C compiler prerequisite absent |
| Two-device LAN, 8 GiB, soak and performance runs | NOT VERIFIED | Need approved runtime and devices |
| Report DOCX/PDF | PASS: artifact checks | Draft, 22 sections, 10 PDF pages; ZIP/parser and PDF image bounds checks |
| Presentation | PASS: artifact checks | Draft, 16 slides, speaker notes, native PowerPoint rendering |
| Diagrams | PASS: generation | Ten diagrams, each Mermaid/SVG/PNG; diagrams describe current boundaries |
| Windows policy attribution | PASS | Selected CodeIntegrity records identify VerifiedAndReputableDesktop/SAC |
| GitHub push | BLOCKED | Connector cannot see repository; no remote commit claimed |
| License | NOT IMPLEMENTED | Owner's choice pending |

## Changes in this checkpoint
Node settings now persist and reload; known peers and recent benchmark results restore on startup. Shutdown cancels outstanding operations and prevents new work from entering. Transfer progress is bounded. Preview reads have a hard size limit. TCP handshakes react to cancellation and readers have a silence timeout. Frontend source is formatted and settings can be saved.

See evidence JSON and screenshots in docs/evidence. Browser screenshots verify frontend rendering only. New Go regression tests were compiled, not run.

## Failed checks and recovery
An invocation of npm resolved to npm.ps1 and failed PowerShell authorization; the installed npm.cmd entry point completed the build. No execution-policy setting was changed. Earlier browser sandbox launch failures were resolved by approved execution of signed Chrome outside the tool sandbox, with Chromium sandboxing retained. Earlier event-log access failures were resolved by an approved read-only query outside the tool sandbox.

## Remaining scope
Current runtime/race/LAN validation, hub-mediated file relay/download, stronger recovery, measured benchmark data and final academic submission remain unfinished. No security control was disabled or bypassed.
