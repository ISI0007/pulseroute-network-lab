# PulseRoute 0.1.0 development preview

**UNSIGNED DEVELOPMENT PREVIEW - not a trusted signed release or setup installer.**

PulseRoute is a Windows x64 Computer Networks laboratory from Zehri Developers, now available under the MIT license. This proposed preview includes portable desktop and CLI executables, license notices, privacy/removal instructions and a build manifest with the exact source commit and executable SHA-256 values.

## Included behavior
- TCP multi-client messaging with correlated server acceptance and socket-write counts.
- Direct streamed file transfer with SHA-256 verification and cancellation.
- Operator-requested LAN discovery and TCP/UDP echo experiments.
- Local SQLite settings, history and recorded experiment samples.

## Validation
Core tests and Linux race detection, Windows desktop backend tests, TypeScript production build, Windows executable builds and consistent product/version metadata are checked in GitHub Actions. A 128 MiB streaming-integrity regression and packet impairment fixtures are included. This is not evidence of native GUI interactions or two-device LAN validation.

The asset is built from commit 2a85693a9e0514ffc971bfef5c452c1e1119c44c in workflow run https://github.com/ISI0007/pulseroute-network-lab/actions/runs/40629168030. Publish only after that run succeeds and asset contents are verified.

## Requirements and limitations
Windows x64 and Microsoft Edge WebView2 are required. Extract the full ZIP before use. Smart App Control may block these unsigned files; do not disable security controls. No encryption or peer authentication; use trusted private networks. File relay through the hub, transfer resume, native GUI checks, two-device LAN, 8 GiB and soak validation are unfinished.

## Code signing policy
See https://github.com/ISI0007/pulseroute-network-lab/blob/main/docs/CODE_SIGNING_POLICY.md
Free signing sponsorship is being explored. No provider approval, sponsored certificate or signed package is claimed. The purpose of this preview is to expose the actual current development build and its limitations, not to satisfy host security policy.

## Installation and removal
This is portable software. Unzip into a writable folder; run only if the machine's policy permits. Read the included READ-ME.md, PRIVACY.md and DISTRIBUTION.md. To remove, close PulseRoute, retain any wanted received files/history, and review/delete its extracted folder and data. A custom PULSEROUTE_DATA directory is separate. Shared WebView2 is not removed. Firewall rules are never added or removed automatically.
