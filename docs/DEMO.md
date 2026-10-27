# Demonstration script
**Precondition:** Use an administrator-approved execution environment. Smart App Control currently blocks the unsigned Go binaries on the development host. Do not disable controls. Native desktop workflow validation remains BLOCKED.

## Six-minute walkthrough
1. Launch PulseRoute on two approved Windows machines with WebView2. Show data directory and peer name in Settings. On the receiver choose its LAN interface (or 0.0.0.0:42191) and matching UDP listener :42192. Save settings, then start the receiver.
2. Connect both clients to the same TCP hub endpoint. Send a message from each; explain server acceptance, socket-write acknowledgment, and the absence of read receipts.
3. Select an image and a video file through the sender's file picker. Queue them and set the receiver's TCP endpoint. Show hashing, streaming, progress, cancellation and final verified status. Video transfer is not a video call.
4. Compare SHA-256 at both ends. Received files have generated prefixes to avoid overwriting existing files. Preview a supported image or a small supported video (16 MiB inline-preview limit).
5. Run TCP and UDP echo experiments with the same peer, 100 probes and 1024-byte payloads. Export raw JSON. Explain RTT, loss, RTT variation and why stop-and-wait goodput is not link capacity.
6. Restart the app. Show restored node settings, known peers, history and saved experiment results. Stop the listener and show explicit connection/recovery state.

## Honest fallback when execution is blocked
Show the source, exact prior test assertions, compiled artifacts and genuine frontend-only screenshots. State that no fresh backend runtime, two-device LAN or throughput result was obtained. Never present the fallback as a successful live networking demonstration.

## Reproducible evidence checklist
Record OS/build, Go version, commit, endpoint addresses, topology, payload/probe count, firewall policy approval, sample timestamps and raw output. Perform at least five runs per condition. Change one variable at a time. Report failures and outliers rather than replacing them with synthetic numbers.
