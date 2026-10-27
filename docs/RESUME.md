# PulseRoute continuation checkpoint
Updated 28 October 2026. The user resumed development; the earlier pause is no longer active.

## Active location
Project: D:\Project\PulseRoute\pulseroute-network-lab.
Toolchain, caches, temporary files, browser dependency and Python libraries: D:\Project\PulseRoute\work.
Go: work\go\bin. GOCACHE: work\gocache. GOPATH: work\modules. TEMP/TMP: work\tmp.
Python document tools: D:\tools\venv\Scripts\python.exe with PYTHONPATH=D:\Project\PulseRoute\work\python-libs.
All project work remains on D:. Do not delete unrelated C: remnants.

## Current state
Networking, SQLite, desktop bridge and six-screen frontend implemented. This checkpoint adds persisted node settings, peer/result restoration, lifecycle cancellation, bounded transfer state, bounded preview reads and connection deadlines. Current source compiles; runtime remains unverified after historical passes.

Draft report DOCX/PDF, 16-slide PPTX with notes, ten diagrams and frontend screenshots are in deliverables and docs. These explicitly state remaining limitations. Native PowerPoint rendering and document structure checks passed.

## Confirmed blocker
CodeIntegrity event 3077 records 14737 (CLI) and 14722 (store test) identify Smart App Control policy VerifiedAndReputableDesktop, GUID {0283ac0f-fff1-49ae-ada1-8a933130cad6}. See WINDOWS_APPLICATION_CONTROL.md and selected JSON evidence. Do not repeat blocked executions without a meaningful policy/signing change. Never disable controls, add bypasses, or infer that conversational permission changes Windows policy. SAC has no individual-app exception; trusted signing or an approved independent development/CI environment is required.

GitHub connector could not access ISI0007/pulseroute-network-lab despite the user's report that access was granted. No remote push succeeded. Recheck only when access changes; preserve the remote README and use a codex/ branch, never force-push. License choice remains unanswered.

## Verification boundary
PASS: final vet checks, internal/desktop test compilation, production builds, frontend six-screen smoke test, Python SQLite schema test, document validation and native slide rendering.
BLOCKED: Go runtime and race tests.
NOT VERIFIED: native desktop workflows, Go SQLite driver/restart tests, two-device LAN, large transfers, soak and performance data.
Frontend screenshots have a bridge-unavailable banner and are not native screenshots. No synthetic timing claims.

## Next work
1. Obtain an approved runtime/signing or CI route and run current Go tests, then fix failures.
2. Restore GitHub connector access, reconcile remote README and push the checkpoint.
3. Exercise settings restart, cancellation/shutdown, reconnect, queue, discovery, previews and two-device transfers.
4. Implement missing hub file relay/download and recovery according to the original specification.
5. Capture reproducible raw benchmarks and native screenshots; revise drafts into a final submission.
6. Obtain the owner's license and academic identity details.

See PROGRESS.md for per-check status. Historical passing tests do not validate current changes.
