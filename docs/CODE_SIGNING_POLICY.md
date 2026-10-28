# Code signing policy

Status: sponsorship preparation only. PulseRoute has not been accepted by SignPath Foundation and no sponsored signature or endorsement is claimed.

## Ownership and responsibilities
The public source repository is ISI0007/pulseroute-network-lab. Zehri Developers is the product brand. ISI0007 is the repository maintainer, reviewer and proposed release-signing approver. Repository access is the authority for code changes; a CI pass does not authorize signing or publishing a release. Each release requires the owner's manual approval. Multi-factor authentication for GitHub and any signing account must be confirmed before enrollment; account-security configuration is reviewed privately with the owner and is not published in repository files.

## What may be signed
Only PulseRoute binaries built from a reviewed commit in this repository, with matching product name/version metadata and passing CI evidence. Do not sign unrelated software, substitute local binaries for CI outputs or attach signatures to modified artifacts. Third-party dependencies retain their licenses and notices. Keys stay with the approved signing service/provider and must never enter source control.

## Verification and distribution
The packaging script requires valid timestamped Authenticode verification before creating a signed distribution ZIP. It records hashes and signer details. Test signed outputs on a protected target computer before describing native behavior or Smart App Control compatibility as verified. No script disables security controls or modifies trust stores.

## Privacy and removal
See PRIVACY.md and DISTRIBUTION.md. Release pages must link this policy and the applicable notices.

## Proposed sponsorship attribution
If SignPath accepts the project, add its required service/certificate attribution and approved configuration before the first signed release. Until approval, do not display a sponsorship badge or claim free signing is provided. Sponsored certificate identity would belong to SignPath Foundation; product branding would remain PulseRoute/Zehri Developers.

## Review and revocation
Investigate reported signature or provenance mismatches, stop distribution of suspect packages, and notify the provider through its official process. Preserve the affected commit, hashes and CI evidence for investigation. Do not automatically revoke certificates or publish reports without the owner's authorization.
