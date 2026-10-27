"""Verify the actual SQLite DDL with Python SQLite; this is not a Go-driver test."""
import json, pathlib, sqlite3, tempfile
from datetime import datetime, timezone
root = pathlib.Path(__file__).resolve().parents[1]
with tempfile.TemporaryDirectory(prefix="pulseroute-schema-") as temp:
    path = pathlib.Path(temp) / "history.db"
    db = sqlite3.connect(path)
    db.execute("PRAGMA foreign_keys=ON")
    db.executescript((root / "internal/store/migrations/001_initial.sql").read_text())
    db.execute("PRAGMA user_version=1")
    db.execute("INSERT INTO settings(key,value) VALUES(?,?)", ("node", '{"name":"persisted"}'))
    text = "'; DROP TABLE events; --"
    db.execute("INSERT INTO events(kind,reference_id,created_at,payload) VALUES(?,?,?,?)",
               ("message", "check", "2026-10-28", json.dumps({"text": text})))
    db.commit()
    db.execute("BEGIN")
    db.execute("INSERT INTO benchmark_sessions VALUES(?,?,?,?,?)",
               ("run", "udp", "127.0.0.1:42192", "2026-10-28", "{}"))
    try:
        db.execute("INSERT INTO benchmark_measurements VALUES(?,?,?)", ("missing", 0, "{}"))
    except sqlite3.IntegrityError:
        db.rollback()
    else:
        raise AssertionError("foreign key accepted an orphan measurement")
    assert db.execute("SELECT COUNT(*) FROM benchmark_sessions").fetchone()[0] == 0
    assert json.loads(db.execute("SELECT payload FROM events").fetchone()[0])["text"] == text
    assert db.execute("PRAGMA integrity_check").fetchone()[0] == "ok"
    db.close()
    db = sqlite3.connect(path)
    assert db.execute("PRAGMA user_version").fetchone()[0] == 1
    assert json.loads(db.execute("SELECT value FROM settings WHERE key='node'").fetchone()[0])["name"] == "persisted"
    db.close()
result = {"status": "PASS", "scope": "SQL schema only, executed with Python sqlite3; Go driver and application integration remain BLOCKED",
          "checks": ["DDL creation", "parameterized text", "foreign key enforcement", "transaction rollback", "reopen persistence", "integrity_check"],
          "sqliteVersion": sqlite3.sqlite_version, "recordedAt": datetime.now(timezone.utc).isoformat()}
(root / "docs/evidence/schema-check.json").write_text(json.dumps(result, indent=2), encoding="utf-8")
print(json.dumps(result, indent=2))
