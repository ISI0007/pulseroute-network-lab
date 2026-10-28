# PulseRoute Windows distribution

Status: trusted signing is pending an approved publisher identity. Existing CI executable artifacts are unsigned. Do not describe them as signed or as a setup installer.

## Signing prerequisites
Use an existing publicly trusted code-signing identity or an approved remote signing service. The publisher completes identity checks with the provider. No certificate was found in this computer's CurrentUser/My or LocalMachine/My code-signing stores on 28 October 2026; SignTool was also absent. No key was exported, trust store changed, service enrolled or money spent.

Microsoft Artifact Signing is an option only where the publisher is eligible. Public Trust individual developers currently require US/Canada residency; organization eligibility is broader. Other locations need an eligible trusted CA/provider. Private Trust and self-signed certificates are not substitutes for public distribution. Do not send private keys, passwords or identity documents in chat or commit them to GitHub.

## Concrete packaging process
1. Select a successful CI run and download the Windows executable artifact for its exact commit. Review and approve those binaries for signing.
2. Install SignTool from the official Windows SDK on the approved signing host. Use the existing certificate/provider's supported key integration; do not export the private key.
3. Run scripts/package-signed.ps1 with SourceCommit (full 40-character SHA), InputDirectory, SignToolPath and the approved CurrentUser/My CertificateThumbprint. It signs copies with SHA-256 and an RFC 3161 timestamp, verifies both executables with Authenticode policy, then emits a portable ZIP with publisher/signature details and SHA-256 hashes. Original binaries are preserved.
4. For binaries already signed by a hardware/cloud service, use VerifyOnly instead of CertificateThumbprint. Valid timestamped signatures and SignTool verification remain mandatory. No ZIP is emitted on failure; incomplete staging folders are not distributable packages.
5. Inspect the publisher and verify on a separate protected Windows machine before distribution. A locally valid signature alone cannot establish public trust on another PC or guarantee SmartScreen reputation or native application behavior. No GitHub release is published automatically.

## Using the verified portable package
Requires Windows x64 and the official Microsoft Edge WebView2 runtime. Extract all files into a writable folder; run PulseRoute.exe. Local history/received files are stored under the application's data directory. This is a portable package, not a setup installer.

Use the same trusted private LAN. Start a receiver on 0.0.0.0:42191 (TCP) and 0.0.0.0:42192 (UDP experiments). Connect to its actual LAN IPv4 address, not 0.0.0.0 or another machine's 127.0.0.1. Discovery uses UDP 42190. Let the operator approve appropriate private-network firewall access; the application does not add rules automatically. Both chat clients connect to the same hub; direct file transfers target the receiver's TCP endpoint. Enable a receiver on each computer when sending files in both directions.

No encryption or peer authentication; use trusted networks. Native UI, two-device LAN, 8 GiB and soak validation remain pending. Read docs/PROGRESS.md and the CI verification addendum for verified scope.

## Official references
- https://learn.microsoft.com/en-us/windows/apps/develop/smart-app-control/code-signing-for-smart-app-control
- https://learn.microsoft.com/en-us/windows/win32/seccrypto/signtool
- https://learn.microsoft.com/en-us/azure/artifact-signing/quickstart

## Publisher enrollment
Zehri Developers is the product brand; the certificate publisher must be the identity independently validated by the signing provider. No verified certificate publisher identity is claimed yet.

Two routes require the owner's decision before enrollment:
- Individual commercial certificate: Certum Standard Code Signing and SSL.com IV describe individual validation. The individual certificate normally displays the validated person's name. Confirm the applicant's eligibility, accepted identification/address documents, total certificate/cloud-signing fees and refund terms directly with the provider before paying. Submit identity documents only through the provider's official verification channel.
- Open-source sponsorship: SignPath Foundation is a possible application route if the owner chooses an OSI-approved license and meets its other project/reputation/release conditions. Approval is not automatic, and the certificate publisher is SignPath Foundation. The owner approved MIT on 28 October 2026 and LICENSE is present. No sponsorship application has been submitted.

Microsoft Artifact Signing Public Trust has geographic and publisher eligibility restrictions. Check its current eligibility list before enrolling. A Private Trust profile is not a substitute for public distribution.

Official pages:
- https://support.certum.eu/en/code-signing-required-documents/
- https://www.ssl.com/products/software-integrity/code-signing/
- https://signpath.org/terms

Suggested provider inquiry (not sent): Can you issue a publicly trusted Windows Authenticode code-signing certificate to me as an individual developer? Which legal name will be shown, what documents are accepted, and what is the complete cost including secure key storage and signing access? Supply location and identity details privately to the provider when required.

## Free sponsorship preparation
The owner selected exploration of free open-source sponsorship and approved MIT. See SPONSORSHIP_APPLICATION_DRAFT.md for the current draft/checklist, CODE_SIGNING_POLICY.md for release responsibilities and PRIVACY.md for data/removal behavior. No sponsored certificate has been granted.
