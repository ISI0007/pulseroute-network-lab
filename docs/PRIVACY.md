# PulseRoute privacy and local data

PulseRoute has no application analytics endpoint, advertising integration, registration server or cloud history service. Its Go networking code sends messages/files to operator-selected TCP endpoints, sends echo probes when requested, and uses LAN discovery on the designated UDP port when the operator requests discovery. Connected sessions exchange heartbeats. Other connected hub clients receive broadcast chat messages.

Local SQLite history can contain message content, peer names/IP endpoints, file names/paths, hashes, timestamps and experiment samples. Received files, temporary transfers, settings and WebView2 state are stored locally. History and transferred content are not encrypted by this application. The operator is responsible for access to the data directory and the trusted LAN. Sharing diagnostics or experiment artifacts is a separate deliberate action; review them for private content first.

The embedded Windows web runtime is Microsoft Edge WebView2. Its platform behavior, browser security features and updates are governed by Microsoft, separately from PulseRoute's networking implementation. Installing or updating that runtime can contact Microsoft. See https://learn.microsoft.com/en-us/microsoft-edge/webview2/concepts/security and https://privacy.microsoft.com/privacystatement.

The signing/build services receive source or build artifacts when the maintainer deliberately uses them; they are not required for normal peer messaging after installation. No private signing keys or enrollment identity documents should be collected by PulseRoute or committed to this repository.

## Removing the portable application
Close PulseRoute on each computer. Preserve any received files/history you want to keep. Delete the extracted portable application folder and its data subfolder when no longer needed; this removes local records and received files stored there. If PULSEROUTE_DATA was configured, that separate directory must be reviewed and removed separately by its owner. WebView2 is a shared prerequisite and is not removed by deleting PulseRoute. Any manually approved firewall rule can be reviewed in Windows settings; PulseRoute does not change those rules automatically.
