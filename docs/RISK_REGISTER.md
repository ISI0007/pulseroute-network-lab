# Risk register
| Risk | Effect | Mitigation/status |
|---|---|---|
| Repository inaccessible | Cannot push commits | Retry connector; retain local source and history |
| Slow toolchain download | Cannot execute Go checks | Resume official archive; verify published SHA-256 |
| No C compiler | Windows race detector unavailable | Linux CI plus explicit NOT VERIFIED locally |
| One physical device | LAN/firewall behavior unknown | Separate loopback evidence from later two-device run |
| Unauthenticated sockets | Spoofing, disk exhaustion | Trusted LAN scope, explicit listener start, bounds/timeouts; authentication deferred |
| Crash between rename and SQL | Orphan file | Record limitation; add reconciliation before durable-recovery claim |
| Sender loses final ack | File may exist despite reported failure | Show uncertain result; no automatic retry |
| No resume | Interrupted large file restarts | Temporary cleanup and explicit manual retry |
| UI depends on WebView2 | Launch can fail on clean Windows | Doctor/preflight and packaging instructions |
| License pending | Distribution terms undefined | Do not add LICENSE without owner choice |
