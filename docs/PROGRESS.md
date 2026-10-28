# Progress and verification
Updated 28 October 2026. Active project: D:\Project\PulseRoute\pulseroute-network-lab.

**Working development checkpoint; final native demonstration and two-device validation remain unfinished.**

## Verified CI milestones
Latest workspace: [run 40669618323](https://github.com/ISI0007/pulseroute-network-lab/actions/runs/40669618323), source 8d267be864a07642ea0d23090a9c1b919f451104: Linux, Windows and native-smoke jobs succeeded. Native result, screenshot, artifact digests and provenance are under evidence/native-8d267be; five desktop backend pass records are retained under evidence/ci/8d267be-windows.

Recorded experiment baseline: [run 40626295936](https://github.com/ISI0007/pulseroute-network-lab/actions/runs/40626295936), source 4ef6e0c, retains the independently recalculated twenty loopback runs under evidence/ci/4ef6e0c.

| Check | Status | Evidence and scope |
|---|---|---|
| Linux core tests with race detector | PASS | 27 test/subtest pass records, zero failures; go test -json -race ./... -count=1 |
| Windows core tests | PASS | Same 27 test/subtest outcomes; includes Go SQLite migrations/settings/history and LAN discovery over loopback |
| Windows desktop backend tests | PASS | Five tests: settings restart, bounded progress, concurrent shutdown, persisted message confirmation and four-palette theme validation/restart |
| Streamed 128 MiB file | PASS | Streaming source generation and received SHA-256/length verification; no whole-file buffer; no speed improvement claimed |
| Active cancellation and failed-history cleanup | PASS | Receiver partial/destination cleanup asserted after shutdown |
| Concurrent confirmed messages | PASS | 16 correlated acknowledgments with two client sessions; rejection and pre-send cancellation tests |
| UDP impairment handling | PASS | Actual datagram fixtures exercise delayed, duplicated, altered and missing replies plus cancellation |
| Static analysis and production builds | PASS | Linux vet; desktop vet; TypeScript/Vite; Windows CLI and Wails builds in CI |
| Real loopback experiments | PASS | 20 runs, 100 probes/run, 1024-byte payload; Windows/Linux TCP/UDP; 2000/2000 unique replies |
| Independent metric recalculation | PASS | Raw samples reproduce received count, loss, duplicates, ordering, mean RTT, jitter and echo goodput |
| Frontend themes/layout/accessibility checks | PASS | Six screens, four themes, 72 layouts, 32 representative contrast pairs (minimum 5.425:1), keyboard/reduced-motion checks; no mocked bridge |
| Local Go execution | BLOCKED | SAC now also blocks go.exe; event 3077 record 14769; no repeat after block |
| Native Wails window/receiver/settings smoke | PASS: limited scope | Successful run 40669618323: actual native window, navigation actions, UI-started receiver with real TCP ping, Ocean theme selection persisted in SQLite, settings save/stop and graceful shutdown |
| Native file dialogs, drag/drop and media workflows | NOT VERIFIED | Native smoke does not cover complete file/preview workflows |
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

The owner-approved unsigned development preview v0.1.0-preview.1 is published. Its public ZIP and checksum were checked, with eight expected portable files and a verified source manifest. This does not change trusted-signing status; provider acceptance remains pending.

## Sponsorship application
Submitted with explicit owner authorization. The owner reported a website thank-you/confirmation message; automation did not independently capture the final acknowledgement. Provider approval, account activation and trusted signing remain pending. No contact details are included in repository files and optional marketing was not selected.

## Workspace refresh milestone
The frontend now provides Forest, Midnight, Daylight and Ocean palettes, defaulting fresh installations to Forest while preserving saved choices. Refresh scheduling is idle/visibility-aware, with no measured energy-saving claim. Full build and native baseline passed in run 40668188066 at 755818d. See UI_DESIGN.md, current UI evidence and the retained native screenshot/result/provenance. The existing published unsigned preview predates the redesign; current artifacts are separate and no release was replaced.

The final accessibility correction delegates modal keyboard focus to the browser's native dialog behavior so video controls are not skipped by a custom Tab trap. Native Ocean selection and database persistence passed in the final run; all five Windows desktop backend tests passed. The final portable build is saved on D: at work/verified-8d267be-UNSIGNED with archive/executable hashes checked. No new public release was published.
