# Windows execution-block diagnosis
Recorded 26 October 2026, approximately 13:31?13:34 Asia/Shanghai.
Scope: read-only investigation. No security controls, signatures, trust stores, firewall rules or application-control policies were changed.

## Conclusion
**Smart App Control is enabled in enforcement mode and is the leading explanation. Exact attribution of the blocked launches is NOT VERIFIED.**
Observed registry value:
`HKLM\SYSTEM\CurrentControlSet\Control\CI\Policy\VerifiedAndReputablePolicyState = 1`.
Microsoft documents 1 as On (Enforcement). SAC is built on Windows application-control/code-integrity technology, so its error wording overlaps with App Control for Business/WDAC.

Both built PulseRoute PE executables are unsigned (Get-AuthenticodeSignature returned NotSigned). That is evidence about signing, not proof that a binary is safe or the sole cause of its block.
The AppLocker SrpV2 registry path was absent. This makes locally configured AppLocker less likely but does not rule out all managed policies.
Code Integrity and AppLocker event-log reads failed with **Attempted to perform an unauthorized operation.**
DeviceGuard CIM query failed with **Access to a CIM resource was not available to the client.**
Read-only `CiTool.exe -lp -json` returned `{"OperationResult":-2147024891}` (0x80070005 / access denied).
Therefore no event record ID, policy GUID/name or exact per-launch event ID was obtained. Do not describe candidate IDs below as observed events.

## Exact blocked commands and files
Working directory: `D:\Project\PulseRoute\pulseroute-network-lab`.

1. `go test ./... -count=1 -v`
   Go executable: `D:\Project\PulseRoute\work\go\bin\go.exe`.
   The compiler ran; child test execution was blocked:
   - `D:\Project\PulseRoute\work\tmp\go-build1965270774\b172\bench.test.exe`
   - `D:\Project\PulseRoute\work\tmp\go-build1965270774\b186\network.test.exe`
   - `D:\Project\PulseRoute\work\tmp\go-build1965270774\b188\store.test.exe`
   - `D:\Project\PulseRoute\work\tmp\go-build1965270774\b190\wire.test.exe`
   Exact Go error: `fork/exec <path>: An Application Control policy has blocked this file.`
   Earlier blocked invocation used go-build3774514620 with the same package executable names.
   Temporary test binaries may have been cleaned by Go; do not assume they still exist.

2. `& '.\build\bin\pulseroute-cli.exe' -mode selftest -data '..\work\experiment-data' -count 100 -bytes 1024`
   Exact file: `D:\Project\PulseRoute\pulseroute-network-lab\build\bin\pulseroute-cli.exe`.
   Exact PowerShell error: `Program 'pulseroute-cli.exe' failed to run: An Application Control policy has blocked this file`.
   FullyQualifiedErrorId: `NativeCommandFailed`.
   No raw benchmark output was produced.

The desktop `build\bin\PulseRoute.exe` compiled and was found unsigned, but was **not launched**. It must not be reported as a separately observed blocked executable.

## Relevant event IDs to inspect (not observed here)
- CodeIntegrity/Operational **3077**: enforced application-control block, including SAC.
- Correlated **3089**: signature information, useful for understanding unsigned/untrusted status.
- **3099**: policy activation information; correlate policy identity.
- AppLocker/EXE and DLL **8004**: an executable/DLL was prevented from running.
- CodeIntegrity **3076** is audit/evaluation, not an enforced block.

An administrator should use Event Viewer to filter the relevant logs around 26 October 2026, 13:17?13:30 +08:00 and search for the exact paths above. Retain XML details (policy ID/name, activity ID, file hash, process path and timestamp). The helper script scripts/collect-policy-evidence.ps1 performs read-only collection and does not rerun the program.

## Appropriate approval process
1. Review the source, provenance and build output; use docs/evidence/executable-sha256.json to identify the saved artifacts. These are ordinary SHA-256 file hashes, not automatically the Authenticode/PE hashes used by WDAC policy rules.
2. Confirm the enforcing policy from the event and policy inventory before requesting a rule change.
3. **If SAC is the enforcement source:** Microsoft documents no per-app exception. The supported developer route is a valid code-signing certificate recognized by SAC, followed by testing the signed build. An administrator's verbal approval or a locally generated self-signed certificate does not automatically satisfy SAC. No certificate purchase, credential handling, trust-store installation or signing was performed.
4. **If organization-managed WDAC is responsible:** the policy owner may approve a narrow supplemental policy using a reviewed file hash or trusted publisher/signer under the organization's change process. Generate rule hashes using the supported policy tools; a plain file SHA-256 is not a substitute for the policy's PE hash. New builds change hashes. Avoid broad allow rules for writable temporary directories.
5. **If AppLocker is responsible:** the administrator may add a narrowly scoped allow rule for the reviewed executable hash, or an appropriate publisher rule after signing, then verify effective policy and deny-rule precedence.
6. Runtime tests can also be run in an explicitly approved development/CI environment under its own policy. Do not change this host's controls to make the tests pass.

## Other failures, kept separate
- Earlier C: storage exhaustion and temporary-file rename errors were resolved by moving active work to D:; core tests then passed.
- Race detection: `go: -race requires cgo; enable cgo by setting CGO_ENABLED=1`; no C compiler is installed. This is a toolchain prerequisite, not the SAC diagnosis.
- Edge/Chrome headless automation failed with crash-server/COM access errors. No Code Integrity evidence attributes these browser failures to SAC.
- Frontend Vite realpath failure was resolved by preserving project symlink paths during bundling; TypeScript and production builds pass.
- GitHub target remains inaccessible via connector (404); no remote commits are claimed.

## Primary references
- [SAC modes and events](https://learn.microsoft.com/en-us/windows/apps/develop/smart-app-control/test-your-app-with-smart-app-control)
- [SAC per-app exceptions and code signing](https://support.microsoft.com/en-us/windows/security/threat-malware-protection/smart-app-control-frequently-asked-questions)
- [App Control troubleshooting and correlation](https://learn.microsoft.com/en-us/windows/security/application-security/application-control/app-control-for-business/operations/appcontrol-debugging-and-troubleshooting)
- [WDAC rule levels and PE hashes](https://learn.microsoft.com/en-us/windows/security/application-security/application-control/app-control-for-business/design/select-types-of-rules-to-create)
- [AppLocker event IDs](https://learn.microsoft.com/windows/security/application-security/application-control/app-control-for-business/applocker/using-event-viewer-with-applocker)
