# Software Requirements Specification
## Purpose and actors
PulseRoute supports a student operator demonstrating real transport behavior, a peer sending or receiving data, and an evaluator reproducing experiments.

## Functional requirements
FR01: start/stop a bounded concurrent TCP listener.
FR02: connect multiple named clients and broadcast text with explicit acceptance acknowledgment.
FR03: reconnect after a socket interruption; visibly disclose uncertain delivery.
FR04: transfer regular binary files, including images/videos, without loading the file into RAM.
FR05: validate size and SHA-256 before recording success.
FR06: display progress, speed, ETA, completion and failure; cancel active transfers.
FR07: identify LAN peers and connect directly.
FR08: compare TCP/UDP echo probes and retain raw sequence/RTT observations.
FR09: persist settings, peers, messages, file metadata, transfers and benchmark results in SQLite.
FR10: provide desktop dashboard, messaging, transfer queue, preview, history, settings, and errors.
FR11: export measured evidence and provide a reproducible demo.

## Nonfunctional requirements
NFR01: 128 KiB maximum frame including type; 64 KiB file chunks; 8 GiB declared file limit.
NFR02: at most 64 simultaneous TCP handlers. Socket deadlines and cancellation bound stalls.
NFR03: remote names never select arbitrary destinations. No overwrite of existing received files.
NFR04: parameterized SQL, explicit migrations, transaction boundaries, indexed chronology.
NFR05: dark/light accessible UI with empty, busy, failed and success states.
NFR06: Windows executable build and tests are recorded separately from source implementation.

## Acceptance and exclusions
Byte-for-byte comparison validates transfers; eight clients validate basic concurrency. Loopback results are not LAN results. SHA-256 is integrity checking, not authentication. No promise of exactly-once durable recipient delivery, offline messages, resume, encryption, NAT traversal or live video calling unless separately implemented and tested.
