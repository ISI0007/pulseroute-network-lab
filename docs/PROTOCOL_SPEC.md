# PulseRoute v1 protocol
## TCP envelope
Four-byte unsigned big-endian length N, then one-byte type and N−1 payload bytes. Require 1 ≤ N ≤ 131072 before allocation. Types: 1 JSON control, 2 binary chunk, 3 end. JSON controls carry version=1 and an operation. Unknown versions are rejected.

## Chat
Client sends hello(peer), server sends welcome. Client sends message(id,text), server validates ID length and text ≤8192 bytes and records acceptance. It writes the message to currently connected sessions, then ack(id,count). Count means successful socket writes, not durable receipt or human reading. Slow recipients have a five-second write deadline and are disconnected. Heartbeats every 20 seconds keep the 90-second read deadline active. Offline replay and duplicate suppression are not implemented; manual retries can duplicate text.

## File upload
put(id,name,size,sha256) → ready → chunk* → end → complete(id,size,sha256).
The sender hashes in a preliminary streaming pass. The receiver enforces the declared size, chunk limit, and digest. A changed source file fails verification. Empty files send no chunks. File IDs use 128 random bits. Received filenames have a separate generated prefix. Interrupted/corrupt data remains temporary and is removed on handler exit. No transfer resume is claimed.

## Limits and safety
64 concurrent handlers; 15-second initial handshake deadline; 30-second per-file-operation deadline; maximum 8 GiB. ASCII filenames of at most 120 characters; separators, colon, trailing dot/space and consecutive dots rejected. Generated prefixes prevent Windows reserved-name collisions. The destination is an operator-controlled private directory, not a remotely chosen path. Encryption and peer authentication are not present; deploy only on trusted networks.

## Experiments
ping/pong provides a one-shot TCP RTT. echo provides bounded binary request/response probes. UDP sequence protocol and formulas are documented alongside the actual implementation when verified.

UDP echo payloads contain PLS1 magic, a big-endian uint32 sequence, an eight-byte random per-run cookie and padding. Replies must match the size, magic, cookie and padding. Unrelated or altered UDP replies are ignored until the probe deadline; altered TCP replies fail the experiment. The cookie isolates experiments and does not authenticate a peer.
