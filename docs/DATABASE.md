# Database design
SQLite schema version 1 is embedded from internal/store/migrations/001_initial.sql. Open uses busy_timeout=5000, WAL and foreign_keys=ON. Migration and user_version update run in one transaction; versions newer than supported are rejected. One database connection serializes local operations.

| Table | Key | Stored information |
|---|---|---|
| settings | key | Theme plus validated node settings JSON |
| peers | address | Known name and last-seen timestamp |
| events | seq | Kind, reference ID, timestamp, JSON event payload |
| benchmark_sessions | id | Transport, address, timestamp and result JSON |
| benchmark_measurements | session_id + sequence | Raw observation JSON, FK to session |

Messages, transfer metadata, paths, sizes, digests and errors use typed event payloads rather than separate normalized message/file tables. This is a deliberate initial schema, not a claim of a fully normalized media catalog. The events indexes support kind/time and reference lookups. Benchmark session and sample insertion is transactional.

Startup restores validated node settings, up to 128 known peers and the newest 20 benchmark sessions in display order. Known peers are not necessarily online. Snapshot history contains the newest 200 events; received-file preview lookup examines up to 500.

The Python schema check executes the actual DDL and verifies rollback, foreign keys, text parameters, reopen persistence and integrity_check. That PASS does not validate the modernc Go driver or application lifecycle; those runtime tests remain BLOCKED.
