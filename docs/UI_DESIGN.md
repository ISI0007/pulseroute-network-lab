# Workspace design and frontend verification

Updated 28 October 2026. This refresh implements the owner's request for a more professional, friendly interface with calmer colors and more economical idle behavior. It does not claim measured energy savings or environmental certification.

## Appearance
| Theme | ID | Intent | Lowest representative text contrast checked |
|---|---|---|---:|
| Forest | forest | Muted green, warm accent; fresh-install default | 6.563:1 |
| Midnight | dark | Charcoal and mint | 6.628:1 |
| Daylight | light | Paper surfaces and deep green | 5.425:1 |
| Ocean | ocean | Deep blue and sky accent | 6.330:1 |

An existing saved theme is preserved. Native selections use SetTheme and SQLite. Browser-only appearance preview is explicitly labeled and does not pretend to save native settings or provide networking. Genuine operation errors remain visible.

## Interface changes
Clearer sidebar/header hierarchy, selected navigation, session/transfer status, connection topology, empty states and context actions. Settings presents four visual palette cards. Controls have consistent sizing and keyboard focus indicators. Queue/progress controls and navigation have accessible names. Preview-dialog code supports Escape, focus containment, focus restoration and an inert background; complete native media/dialog behavior still needs actual workflow verification.

## Adaptive refresh
The frontend schedules one Snapshot request at a time, at about 500 ms when a connection/listener/transfer is active and 1800 ms while idle. It stops scheduling while hidden and requests fresh state when visible. Missing native bindings are polled lightly for late availability rather than displayed as repeated application errors. Polling changes do not cancel backend transfers or connection heartbeats. Reduced-motion styling follows the operating-system/browser preference.

## Executed frontend checks
Production TypeScript/Vite build passed. The signed-Chrome test serves built assets and never mocks Go methods or fills fake network records. It checks six screens, four theme previews and three widths (1440, native minimum 1000, optional browser 768), totaling 72 no-overflow layouts. Thirty-two representative color pairs exceed 4.5:1 (minimum 5.425:1), with keyboard navigation, visible focus, selected navigation, reduced motion and no JavaScript page errors. This is not a complete WCAG audit. Native theme persistence is covered separately by backend/CI tests, not inferred from CSS preview success.

Evidence: evidence/ui-checks.json, evidence/frontend-*.png and evidence/theme-*-overview.png / theme-*-settings.png. Run scripts/inspect-ui.cjs against the production build with Playwright and installed Chrome. Real file selection/drag-drop/modal media workflows remain outside this browser-only check.

## Native and LAN validation
The native CI job launches the built Wails executable in an isolated application-data directory, invokes navigation/receiver/settings through Windows UI Automation, checks the resulting TCP listener with a real protocol ping, and inspects settings persisted after graceful shutdown. It has no remote-debugging endpoint and changes no security controls. Read its result JSON before claiming PASS; unavailable desktop/runtime prerequisites are explicitly BLOCKED.

Two-computer procedure: LAN_TEST_GUIDE.md, with LAN_ACCEPTANCE_TEMPLATE.csv initially marked NOT RUN. No new release is published automatically. The existing unsigned preview predates this frontend redesign.

## Executed native baseline
Run [40668188066](https://github.com/ISI0007/pulseroute-network-lab/actions/runs/40668188066), source 755818d, passed Linux, Windows and native-smoke jobs. The actual Wails executable created a window; navigation, receiver/start-stop/settings actions were invoked through Windows UI Automation. The UI-started listener answered a real protocol ping, saved settings were read from SQLite after shutdown, and the native screenshot/result were retained with matching artifact checksum under evidence/native-755818d. The hosted screenshot captures the visible runner viewport and is evidence, not a full-window product export. The final native palette-selection test passed in run 40669618323 (8d267be): UI Automation selected Ocean and a separate SQLite read after graceful closure confirmed the persisted value. Final native evidence and provenance are under evidence/native-8d267be.

The initial hidden-window harness run was BLOCKED by lack of an accessible window; visible launch in the isolated CI session corrected that prerequisite. A follow-up harness used an unsupported Invoke pattern for a theme toggle; this was corrected to support the Windows Toggle pattern. Neither change disables security policy or alters production runtime controls. Native file/media dialogs still require separate workflow validation.

The media preview now uses HTML dialog.showModal() for browser-managed focus, Escape cancellation and trigger-focus restoration. This fixes the custom Tab trap that could skip native video-control shadow elements. Full native file selection and media playback workflows remain NOT VERIFIED. Final frontend build/QA and the complete native/core CI run passed after this source correction.
