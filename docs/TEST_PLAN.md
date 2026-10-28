# Test plan
Unit: framing short writes, binary bytes, truncated headers/payloads, oversized/zero lengths; fuzz arbitrary frame input.
Integration: eight TCP clients, broadcast and ack count, disconnect/reconnect, bounded shutdown.
Transfer: zero/one-byte, chunk boundary, 16 MiB, duplicate filenames, content equality, wrong digest, traversal, interruption cleanup, cancellation.
Persistence: migrations twice, reopen, chronological query limits, parameterized special characters.
Experiments: synthetic duplicates/loss/out-of-order metrics, actual loopback TCP/UDP, recorded raw samples.
Desktop: build, launch, actions through real backend, keyboard/input error states, file selection and progress.
System: race detector, go vet, Windows build, large-file/soak tests and two-device LAN when available.
Never promote a test to PASS because it is merely written. Retain exact command, exit result, date and relevant environment.

## Executed CI milestone (28 October 2026)
Run 40626295936 / source 4ef6e0c passed on hosted Linux and Windows. Includes a streamed 128 MiB file, cancellation after the receiver is ready, history-failure cleanup, prompt handshake cancellation, concurrent message confirmation, actual UDP loss/duplicate/delay/altered-payload fixtures, and concurrent desktop shutdown. See CI_VERIFICATION.md and retained JSONL test names for exact scope. Native UI and two-device tests remain pending.

## Workspace refresh coverage
Four palette selections, 72 real frontend layouts and representative contrast/keyboard/reduced-motion checks pass without mocking the backend. Desktop theme validation/restart tests and native window/receiver/settings smoke pass in CI. See UI_DESIGN.md for scope, exact runs, harness failures/corrections and remaining native media/dialog checks. Two-device acceptance steps and NOT RUN template are in LAN_TEST_GUIDE.md and LAN_ACCEPTANCE_TEMPLATE.csv.
