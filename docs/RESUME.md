# PulseRoute continuation checkpoint
Updated 28 October 2026 after successful CI runtime validation.

## Active locations
Project: D:\Project\PulseRoute\pulseroute-network-lab.
Tools/caches/temp/browser/Python libs: D:\Project\PulseRoute\work.
Documents Python: D:\tools\venv\Scripts\python.exe with PYTHONPATH= D:\Project\PulseRoute\work\python-libs and MPLCONFIGDIR=work\matplotlib.
Keep project work on D: and preserve unrelated files.

## Verified state
Core networking, SQLite, discovery and experiments passed on Linux (with race detection) and Windows. Five Windows desktop backend tests passed for the refreshed workspace. Latest source 8d267be; run 40669618323 passed all three jobs, including native window/receiver/settings/theme selection. Native evidence: docs/evidence/native-8d267be. Raw experiment baseline remains docs/evidence/ci/4ef6e0c. Twenty actual loopback runs were independently recalculated; results in docs/experiments/4ef6e0c.

Latest changes: persisted settings/history reload, bounded state/preview reads, lifecycle cancellation, 128 MiB streamed integrity regression, active cancellation cleanup, UDP cookie validation/impairment tests, and correlated message confirmations. UI displays server acceptance/socket-write count and never claims human reading. Source and evidence are pushed to main and codex/pulseroute-foundation.

## Local restriction
Smart App Control VerifiedAndReputableDesktop blocks local Go test/CLI launches and now go.exe itself. Compiler event 3077 record 14769 on October 28; evidence JSON retained. Do not retry blocked commands without a meaningful authorized signing/policy change. Do not disable controls, add exclusions or bypass policy. Use GitHub Actions for Go runtime/static/build validation. gofmt and signed Chrome frontend smoke tests remained permitted. Root Linux race tests and Windows desktop backend tests run in CI. Conversational approval does not change Windows policy.

## Remaining work
1. Native window/navigation/receiver/settings smoke passed in CI run 40668188066. Continue real file dialogs, drag/drop, queue/cancel and media preview workflow checks; these remain unverified.
2. Two-device LAN and larger/soak experiments; record environments and samples. Existing hosted loopback is not a link-capacity benchmark.
3. Decide whether optional hub file forwarding/download and transfer resume are needed for final scope. Existing direct uploads and hub chat are implemented.
4. Revise report DOCX/PDF and presentation into a final submission using the measured evidence/addendum; existing artifacts are earlier progress drafts.
5. MIT was approved and published. Obtain the remaining academic identity details.
6. Continue free signing preparation using SPONSORSHIP_APPLICATION_DRAFT.md. Windows product/version metadata passed CI run 40628906942; license notices and privacy/signing policies are prepared. Do not claim sponsorship or publish a release until approved.

Workflow runs on main source changes and pull requests, skips documentation-only pushes, and retains raw logs/loopback results. Avoid pushing repeated source changes while checking a run because newer runs cancel older ones. Preserve history; never force-push.

## Approved preview publication
Owner approved publishing unsigned prerelease v0.1.0-preview.1, source 2a85693. Public release and ZIP/checksum are verified. See the updated sponsorship draft for application answers. The owner later authorized application submission and reported the website thank-you acknowledgement. Provider review and any account activation remain pending; no signing request or certificate exists. Contact details remain outside the repository.

## Sponsorship submission update
Owner confirmed the application website displayed a thank-you message after the authorized form submission. The automated session had ended without capturing that final acknowledgement, so confirmation is explicitly owner-reported. Await the provider reply; do not duplicate the application or claim acceptance/signing. The owner must complete any official email verification privately.

## Frontend refresh
Four palettes, accessible navigation/controls, keyboard/reduced-motion behavior and adaptive idle/hidden refresh are implemented. Browser QA passed 72 layout cases and 32 representative color pairs, minimum 5.425:1. Fresh installs use Forest; all four themes validate and survive backend restart tests. Native baseline result/screenshot: docs/evidence/native-755818d. A verified unsigned build was saved at D:\Project\PulseRoute\work\verified-755818d-UNSIGNED without executing it locally. It has no trusted signature and is not the existing public preview artifact.

## Final refreshed build verification
Run 40669618323 / source 8d267be passed Linux race/core, Windows core/backend/build and actual native UI smoke. Native UI selected Ocean and persisted it in SQLite. The custom media Tab trap was replaced by native dialog focus behavior. Browser QA passed again after that correction. Final portable files: D:\Project\PulseRoute\work\verified-8d267be-UNSIGNED; zip: work\ci-artifacts\pulseroute-windows-8d267be.zip. Files were hash-checked, never executed locally. Public preview v0.1.0-preview.1 remains unchanged and older; publishing another release needs approval.

## Calls and nearby checkpoint — 29 October 2026
New implementation and regression tests pushed through f2a5261. Calls uses directed TCP signaling and direct WebRTC; nearby is an explicitly enabled 15-minute browser file portal with QR links. Bluetooth opens the Windows OS assistant only. Local frontend build, 96 layout checks, QR round-trip and synthetic-media WebRTC lifecycle passed. Run 40882087286 PASSED all three jobs: 46 Linux race-tested core cases, Windows core plus eight desktop cases, synthetic-media WebRTC/QR checks, executable builds and native navigation through all eight screens. Evidence: docs/evidence/ci/f2a5261. Hash-verified unsigned binaries: D:\Project\PulseRoute\work\verified-f2a5261-UNSIGNED; not executed locally. See CALLS_AND_NEARBY.md for exact limitations. Keep all project files on D:. No new release or trusted signature exists. Physical device testing requires the owner's devices.
