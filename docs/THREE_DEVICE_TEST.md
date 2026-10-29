# Two Windows PCs and Android: supervised acceptance checklist

All physical results start as NOT RUN. Do not mark a test PASS from CI alone. Use source f2a5261 or later with the same executable hashes on both Windows PCs. Current build: D:\Project\PulseRoute\work\verified-f2a5261-UNSIGNED. This is an unsigned portable development folder, not an installer. Copy the entire folder to a writable location on the second PC using an existing permitted method such as USB. Do not use the older public preview for these new features.

## 1. Launch readiness
Use Windows x64 with WebView2 on both PCs. Use the same trusted Wi-Fi for both PCs and Android. A local security block was already observed on the development PC: do not repeat its blocked launch while the policy/signature is unchanged. On any PC where this build has not been blocked, ordinary launch is subject to that PC's policy. If blocked, record BLOCKED and stop native testing on that device; await a trusted signed build or the policy administrator's approved allowlisting process. Do not choose security bypasses. Record only whether each PC opens the app or its exact block text. Do not post personal files or QR session tokens publicly.

## 2. Shared hub and messages
After both native apps can run, identify PC A's Wi-Fi IPv4 with ipconfig. On A open Settings, set distinct display name PC A, TCP listen 0.0.0.0:42191, UDP echo 0.0.0.0:42192; save and start receiver. On B use display name PC B. In the connection desk on BOTH apps set remote endpoint to A's actual IPv4:42191 and Connect. Never use 0.0.0.0 or A's loopback as B's destination. Approve only normal reviewed private-network firewall access if authorized; do not disable the firewall. Send a unique message each direction and confirm it appears on both windows. Use LAN_TEST_GUIDE.md for direct transfers in both directions and SHA-256 comparisons.

## 3. Voice then video
In Calls confirm the other peer is listed. Start a voice call from A, Accept on B, and grant normal microphone permission if requested. Confirm speech both directions; headphones reduce feedback. Check mute, hangup and a fresh second call. Then place a video call and confirm both real cameras and both voices, camera off/on and hangup. Navigate to another screen during a call and verify the call remains active. Reject one incoming call and confirm no capture starts on the rejecting PC. If a permission or media connection error appears, record the exact message and outcome; do not change browser security flags. WebRTC uses dynamic direct LAN media ports and no TURN service.

## 4. Android QR files
On A open Nearby, select its Wi-Fi interface (not This PC only), optionally choose a harmless text/image file through the native picker, and start sharing. This session grants upload/download access to whoever holds the link for 15 minutes. Use a trusted LAN; HTTP file data is not encrypted. On Android scan the QR and open the link in the browser. Download the selected file, then upload a different harmless file from Android and inspect A's received history. Sender-provided hash verification and receiver-only digest are distinct outcomes. Stop sharing and verify a fresh browser request using the old link cannot access the session. Do not expose active QR/link tokens in public screenshots. Android browser calls are not part of this HTTP file portal.

## 5. Windows Bluetooth assistant
If compatible radios and paired devices are available, open Nearby's Windows Bluetooth transfer assistant and follow Windows Send/Receive instructions. Record completion from Windows and inspect the received file. PulseRoute only opens the assistant; it does not itself report transfer completion. Do not mark PASS merely because the assistant opens. Android support depends on the actual device/OS. Apple AirDrop protocol is not implemented.

## Private results sheet

| Check | PC/device | Result | Observed evidence/error |
|---|---|---|---|
| Native startup | A | NOT RUN / previously BLOCKED if unchanged | |
| Native startup | B | NOT RUN | |
| Chat both ways | A+B | NOT RUN | |
| File transfer + matching SHA-256 | A+B | NOT RUN | |
| Real voice + mute + hangup | A+B | NOT RUN | |
| Real video + camera restart | A+B | NOT RUN | |
| QR download + upload + revoke | A+Android | NOT RUN | |
| Physical Bluetooth transfer | supported peers | NOT RUN | |

Use PASS for observed correct behavior, FAIL for an observed defect, BLOCKED for policy/environment restrictions, NOT RUN otherwise. Keep local IPs, device identities, filenames and tokens private unless explicitly approved for publication.
