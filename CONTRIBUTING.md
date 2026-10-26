# Contributing
Work on a focused branch. Keep transport, application orchestration, persistence and interface separate. Bound every remote allocation, validate inputs, close sockets on cancellation, and document acknowledgment semantics.
Run gofmt, go test ./..., and go vet ./... before committing. Race and Windows desktop checks must be recorded separately. Include reproduction steps for failures and raw data for performance claims. Do not commit local histories, personal files, tokens, build caches or node_modules.
