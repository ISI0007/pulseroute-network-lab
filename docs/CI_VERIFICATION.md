# CI verification and report addendum - 28 October 2026

This addendum supersedes runtime-blocker statements in the earlier DOCX/PDF/PPTX drafts. Local Smart App Control is still enforced; approved GitHub-hosted runners supplied runtime verification.

## Reproducible evidence
- Source: [4ef6e0c](https://github.com/ISI0007/pulseroute-network-lab/commit/4ef6e0c52184e0817ca9b28cac6e7e00b36d7337).
- [Successful verification run](https://github.com/ISI0007/pulseroute-network-lab/actions/runs/40626295936).
- Linux job 122611978201: core race tests, static checks and loopback experiment.
- Windows job 122611977755: core/desktop backend tests, frontend build, static checks, Windows binaries and loopback experiment.
- Downloaded artifact ZIP hashes matched GitHub artifact digests before extraction. Retained JSON/JSONL, artifact IDs and exact hashes: evidence/ci/4ef6e0c.
- Test pass records: 27 Linux core tests/subtests; 27 Windows core; four Windows desktop backend. Zero failure records. These are repeated platform runs, not 58 unique tests.

## Measured experiments
Twenty real runs cover five repetitions of each protocol on each OS. Each run used 100 sequential probes with a 1024-byte payload, totaling 2000 observed unique replies out of 2000 sent. Raw RTT samples are retained. The analyzer independently reproduces all aggregate counters/formulas and exports CSV and PNG/SVG charts.

[Measured results and limitations](experiments/4ef6e0c/RESULTS.md).

Recalculate with Python matplotlib installed:
```powershell
python scripts/analyze-experiments.py --input docs/evidence/ci/4ef6e0c --output docs/experiments/4ef6e0c
```
The CLI produces raw data with -mode selftest -count 100 -bytes 1024 in an approved execution environment. Do not execute locally blocked binaries to reproduce it on this host.

## What this establishes
Real socket behavior under current core tests, 128 MiB file integrity, interruption/cancellation cleanup, Go SQLite persistence, loopback discovery, confirmation correlation/rejection, desktop backend lifecycle and application-echo measurement arithmetic. It does not establish a native GUI demonstration, two-machine discovery/connectivity, 8 GiB transfers, soak reliability, real network capacity or superior performance.

## Draft artifact boundary
The report and deck under deliverables were generated before this runtime milestone and remain historical progress drafts. Read them together with this addendum until the final academic revision. No final-submission status is claimed.

## Refreshed native workspace verification
Run [40669618323](https://github.com/ISI0007/pulseroute-network-lab/actions/runs/40669618323), source 8d267be, passed all three jobs. The refreshed frontend has four palettes and 72 browser layout cases with 32 representative contrast checks. Five Windows desktop backend tests include palette validation/restart. The actual Wails window accepted navigation/receiver/settings/theme actions, started a TCP listener that answered a protocol ping, persisted Ocean in SQLite and shut down gracefully. Exact native JSON/screenshot/provenance: evidence/native-8d267be. Native media/file-picker/drag-drop and two-device checks remain pending. See UI_DESIGN.md for harness corrections and evidence scope.
