# Free code-signing sponsorship application draft

**Not submitted. No sponsorship approval or signed release exists.** Prepared 28 October 2026 for owner review.

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
Release/download URL: an unpublished unsigned preview is being prepared for owner review. Publication requires explicit approval. CI artifacts alone are not asserted to satisfy the provider's existing-release requirement.
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
- PENDING: owner-approved release and download page in the form to be signed.
- PENDING: provider review of project maturity/reputation, eligibility and application.
- PENDING: provider-approved artifact configuration and signing integration after acceptance.
- PENDING: native protected-machine and two-device LAN verification.

## Important expectations
MIT approval is not signing approval. A new project can be rejected or asked to build a verifiable track record. Free sponsorship does not promise a Zehri Developers certificate identity: SignPath Foundation would be the certificate publisher. No publisher enrollment, contractual acceptance, release publication or signing request has been performed.

## Form completion
The public application page embeds an external provider form. The project answers above are a reviewable draft, not a completed account registration. Submit only after the pending release/MFA details are resolved and the owner approves the exact application. Enter contact/personal information directly into the official form and keep it out of repository files.

## Latest preparation verification
Workflow 40629168030 at source 2a85693 succeeded, including Windows product metadata checks, regeneration of notices from the actual build toolchain, unsigned-package rejection and portable artifact assembly. Both jobs passed. Archive SHA-256 and both executable hashes were independently checked before preparing the release draft. No binary was executed locally.
