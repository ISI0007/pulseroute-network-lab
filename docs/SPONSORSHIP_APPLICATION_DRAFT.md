# Free code-signing sponsorship application draft

**Submitted: owner reported the website thank-you/confirmation message on 28 October 2026. Sponsorship approval and signed releases remain pending.** The automated browser did not capture the final acknowledgement. Contact details are not stored in this repository.

Application page: https://signpath.org/apply
Terms: https://signpath.org/terms

## Draft project description
PulseRoute is an MIT-licensed educational Windows desktop application demonstrating multi-client TCP chat, direct streamed file transfer with SHA-256 verification, UDP echo measurements and operator-requested LAN peer discovery. It is built in Go, React/TypeScript and Wails with local SQLite history. It targets Computer Networks laboratory demonstrations. It does not scan for vulnerabilities, perform exploitation, bypass host policy or claim Internet NAT traversal.

Repository: https://github.com/ISI0007/pulseroute-network-lab
Project maintainer/contact handle: ISI0007
Product brand: Zehri Developers
License: MIT (owner approved; root LICENSE).
Build instructions and CI: README.md, .github/workflows/verify.yml.
Current test evidence: docs/CI_VERIFICATION.md and the linked successful workflow.
Windows binaries: successful CI artifacts; currently unsigned development builds.
Release/download URL: https://github.com/ISI0007/pulseroute-network-lab/releases/tag/v0.1.0-preview.1 . The owner approved publication of this explicitly unsigned development preview. Provider acceptance of the project remains pending.
Signing policy: docs/CODE_SIGNING_POLICY.md.
Privacy and removal instructions: docs/PRIVACY.md and docs/DISTRIBUTION.md.
Dependency notices: THIRD_PARTY_NOTICES.txt and docs/dependency-notices.json.
Proposed signing approver: ISI0007, with manual approval of each release.
Contact email and account MFA confirmation: owner supplies privately in the provider's official form; not included in this public draft.

## Readiness checklist
- DONE: owner-approved MIT license and public source/build instructions.
- DONE: signed-package verification gate and negative test for unsigned inputs.
- DONE: current CI evidence for core/race, backend, build and loopback tests.
- VERIFIED: Windows product/version resource generation and metadata assertions passed in run 40628906942, source e6d1f83.
- DONE: signing policy, privacy/removal instructions and upstream notice inventory (41 installed components in the initial inventory; CI regenerates for the actual build toolchain).
- PENDING: signing-provider account activation and its MFA; GitHub account security details are reviewed privately with the owner.
- DONE: owner-approved unsigned development preview with public download, notices, policy link, source commit and checksum.
- SUBMITTED: application sent using the owner-authorized contact details; owner reported the website acknowledgement. Provider review of maturity/reputation and eligibility remains pending.
- PENDING: provider-approved artifact configuration and signing integration after acceptance.
- PENDING: native protected-machine and two-device LAN verification.

## Important expectations
MIT approval is not signing approval. A new project can be rejected or asked to build a verifiable track record. Free sponsorship does not promise a Zehri Developers certificate identity: SignPath Foundation would be the certificate publisher. The owner authorized submission of the application. Required application code-of-conduct and processing consents were accepted; optional marketing was left unchecked. No sponsorship acceptance, certificate issuance or signing request is claimed. The published preview remains unsigned.

## Form completion
The official external form was completed and Submit was clicked once. Automation initially saw a submitting state; the owner subsequently reported a thank-you/confirmation message. Treat this as owner-confirmed submission, not an independently captured receipt or provider acceptance. Any email verification/account activation must be completed through the provider's official flow. Do not resubmit a duplicate unless the provider requests it.

## Latest preparation verification
Workflow 40629168030 at source 2a85693 succeeded, including Windows product metadata checks, regeneration of notices from the actual build toolchain, unsigned-package rejection and portable artifact assembly. Both jobs passed. Archive SHA-256 and both executable hashes were independently checked before preparing the release draft. No binary was executed locally.

## Review status
Awaiting the provider's reply. The application honestly disclosed that the project is new and has no established broad adoption or third-party endorsements. No certificate, account approval, signing key or trusted signed executable is available yet.
