# PulseRoute 0.1.0 development preview

**UNSIGNED DEVELOPMENT BUILD.** This CI artifact has no sponsored signature and is not a setup installer or final tested release. Windows Smart App Control may block it. Do not disable security controls to run it.

Source: https://github.com/ISI0007/pulseroute-network-lab
The exact source commit and executable hashes are in build-manifest.json. Product branding is Zehri Developers; it is not a verified certificate publisher identity. Source is MIT licensed, with upstream notices included.

## Requirements and use
Windows x64 and the official Microsoft Edge WebView2 runtime are required. Extract the complete ZIP into a writable folder. If host policy permits, open PulseRoute.exe. Do not run directly inside the ZIP. Refer to DISTRIBUTION.md for LAN endpoints, private-network firewall approval and remaining validation boundaries.

Core networking, persistence and Windows backend/build tests have passed in CI. Native WebView2 interaction and two-computer LAN behavior have not yet been verified. No final demonstration readiness, physical link capacity or 8 GiB reliability is claimed.

## Removal
Close the app. Back up any received files and history you want. Review/delete the extracted folder and its local data when no longer needed. A custom PULSEROUTE_DATA location is separate. The shared WebView2 runtime is not removed. See PRIVACY.md for storage and network behavior.

## Signing status
Free sponsorship is being explored. No application approval, sponsored certificate or trusted signed release exists. Code signing policy: https://github.com/ISI0007/pulseroute-network-lab/blob/main/docs/CODE_SIGNING_POLICY.md
