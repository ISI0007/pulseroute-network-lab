# Calls and nearby sharing

Calls are app-to-app voice/video between connected PulseRoute Windows clients on a reachable LAN. Connect both clients to the same hub, open Calls, select a peer and place a call. The recipient must accept before media capture begins. Allow camera/microphone access through the normal Windows/WebView prompts. Calls use direct WebRTC without public STUN/TURN; Internet NAT traversal is not promised. Media is not recorded. SDP and ICE remain ephemeral; history stores only call summaries.

Nearby starts an explicit, temporary 15-minute HTTP file-sharing session on the selected interface. Choose outgoing files with the native picker, start sharing, and scan the QR or open the link in another device's browser. Stop revokes the session. Anyone holding the link can access selected files and upload while it is active. HTTP file contents are not encrypted: use only a trusted local network. Browser uploads are limited to 512 MiB. A receiver-only SHA-256 is reported as received, not sender-verified integrity. Mobile HTTP browsers cannot capture call media; the browser portal is for files.

Bluetooth opens the standard Windows Bluetooth File Transfer assistant. Windows handles device discovery, pairing, file selection and completion. Presence of the assistant does not establish that a radio or compatible peer exists. PulseRoute does not track assistant transfer completion. iPhone/iPad use the Wi-Fi browser portal; Apple's AirDrop protocol is not implemented.

## Verification scope

Local PASS: TypeScript/production frontend build; 96 layouts over eight screens/four themes/three widths; representative contrast, keyboard navigation and reduced motion; license inventory (70 components). WebRTC lifecycle test uses real browser peer connections with explicitly synthetic tracks and in-process fixture signaling. It checks bidirectional bytes, acceptance before capture, mute, camera restart, rejection and track cleanup. This does not prove native permission prompts or physical cameras.

New Go signaling, nearby HTTP and desktop regression tests run in GitHub Actions because local Go execution is blocked by Windows security policy. See the Actions result for the exact source revision. Do not treat a queued/running workflow as PASS.

Pending physical checks: two Windows computers with real audio/video; Android, iOS/iPadOS and macOS browser transfers; compatible Bluetooth radio transfers. Native macOS/iOS/Android apps are not provided. Trusted code signing remains pending sponsorship approval. No new public release is authorized by this development checkpoint.
