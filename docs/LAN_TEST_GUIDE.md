# Two-computer LAN acceptance guide

Status: a test procedure, not a passed result. Use only a build each computer's security policy permits. The public preview is unsigned and older than subsequent frontend changes; current CI artifacts identify their exact source commit. If Smart App Control blocks execution, stop and await the approved signing route. Do not disable controls or use bypasses.

## Setup
Use two Windows x64 computers on the same trusted private LAN and the official WebView2 runtime. Avoid a guest network/client-isolation configuration. Record the application commit/hash and Windows versions on both computers. Extract the entire portable package into a writable folder.

Use ipconfig to identify each computer's LAN IPv4 address. The addresses below are examples only: A = 192.168.1.20, B = 192.168.1.21. Replace them with the real values. 127.0.0.1 refers to the computer you are using; 0.0.0.0 is a listener binding, never a remote destination.

The operator may approve private-network firewall access for the reviewed application. Ports: TCP 42191 for chat/files, UDP 42192 for echo experiments, UDP 42190 for discovery. Do not turn the firewall off or add unrestricted public-network rules. Managed devices require their policy owner's normal approval.

## Configure both receivers
On each computer, Settings -> TCP listen address = 0.0.0.0:42191, UDP echo address = 0.0.0.0:42192, a distinct Display name (for example Laptop A / Laptop B), Save node settings, then Start receiver. Confirm receiver status on both. Themes are local preferences: Forest, Midnight, Daylight and Ocean.

## Chat through one hub
Use A as the chat hub. On both computers, set Remote TCP endpoint to A's actual IPv4 address followed by :42191 and click Connect. Send text from A, then B. Both windows should show it, including Unicode text. The sender's server-acceptance count means socket writes to connected sessions, not human reading.

Disconnect B and reconnect, then send a fresh message. Old messages are local history, not guaranteed offline replay. Confirm the existing history survives closing/reopening the app.

## Direct file transfer in both directions
A receiver must be running on the destination. For A -> B, select B's IPv4:42191 as the transfer target, choose files through the real file picker or drag/drop, and start the queue. For B -> A, target A's IPv4:42191. The endpoint field is shared with the next connection/transfer destination: changing it does not move an already established chat connection. Disconnect/reconnect explicitly when changing the chat hub.

Start with empty and one-byte files, an image, and a video file. Then use a larger file with a known SHA-256. Verify the receiver's generated destination filename, byte count and hash against the original, for example with PowerShell Get-FileHash -LiteralPath <file> -Algorithm SHA256. Do not run transferred executables as part of this test. Check image/video preview with supported formats below the 16 MiB preview limit; transfer size and preview size are separate limits.

Queue two files with the same original filename from different folders; received files must not overwrite one another. Cancel an active transfer, verify its incomplete file is removed, and retry as a new transfer. Receiver history must not show a corrupt file as verified. Transfer resume is not implemented.

## Discovery and experiments
Use operator-requested LAN discovery. A discovered peer should identify its actual LAN address. If discovery is unavailable, record the failure and use explicit endpoints for other tests; discovery failure is not a full LAN transfer failure.

Run Network Lab TCP and UDP against the other computer's endpoints, initially 100 probes and a 1024-byte payload. Export raw results. Record which machine was sender/receiver, whether Wi-Fi or Ethernet, approximate signal/link conditions, background load and any deliberate impairment. Echo goodput is not physical link capacity. Do not label one-way latency/loss or OS performance differences from these sequential echo measurements.

## Disruption and closure
During a chat session, briefly stop the remote receiver, restart it and observe the reconnect/error state. During a file transfer, deliberately cancel through the UI before testing a separate actual network interruption. Record cancellation separately from accidental interruption. Do not change Windows security policies to simulate failure.

Stop both receivers and close the windows. Confirm the application exits without hanging. Reopen Settings, verify saved addresses/name/theme and inspect persisted history. Check that background/hidden refresh does not cause stale state after returning to the window.

## Record evidence
Use PASS only for an observed successful check, FAIL for a reproducible incorrect outcome, BLOCKED for policy/environment restrictions and NOT RUN for work not attempted. Include original/received hashes, exported raw samples, exact error text and genuine screenshots. Keep private LAN addresses, messages and personal file paths out of public GitHub evidence unless explicitly approved for sharing.

Remaining checks include 8 GiB, multi-hour soak, true two-device behavior and native picker/drag/drop/media workflows. This guide does not mark them complete.
