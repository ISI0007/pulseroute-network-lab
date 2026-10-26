# Project plan
Start: 26 October 2026, Asia/Shanghai. Target window: 26 October–8 November (14 days). Dates are planning targets, not proof of completion.

| Day | Deliverable | Gate |
|---|---|---|
| 1 | Environment, repository, SRS, architecture | Toolchain and access recorded |
| 2–3 | Framing, concurrent chat, reconnect | Multi-client and interruption tests pass |
| 4–6 | File, image and video-file transfer | Byte identity, malformed input, empty/large files |
| 7 | SQLite history, recovery behavior | Restart persistence, interrupted transfer cleanup |
| 8 | LAN discovery and direct transfer | Loopback plus separate two-device evidence |
| 9–10 | UDP/TCP measurements | Raw samples, formulas, reproducible runs |
| 11 | Wails desktop workflows | Actual backend integration and UI tests |
| 12 | Security, race, build, recovery | No unresolved core correctness failures |
| 13 | Report and diagrams | Claims traced to evidence |
| 14 | Presentation and packaging | Real files, rehearsal, verification report |

## Ordered backlog
P0: framing limits; partial writes; concurrency/shutdown; persisted message acknowledgment; file bounds/hash/name safety; cancellation/cleanup; reproducible CLI.
P1: SQLite migrations and indexes; peer registry; UDP sequence/RTT/loss/jitter; comparison experiments; Wails actions and live state; transfer queue; preview; history/settings.
P2: two-device tests; full-size 8 GiB limit test; sustained-load soak; crash recovery; final document review.
Deferred until core passes: resume, TLS, QUIC, internet P2P, live audio/video.

## External dependencies
GitHub repository is user-reported created; connected access still returns 404. License choice pending. Go download required; C compiler absent. Only one Windows host is available, so two-device LAN verification requires later hardware access.
