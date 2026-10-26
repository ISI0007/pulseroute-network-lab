CREATE TABLE settings(key TEXT PRIMARY KEY,value TEXT NOT NULL);
CREATE TABLE peers(address TEXT PRIMARY KEY,name TEXT NOT NULL,last_seen TEXT NOT NULL);
CREATE TABLE events(seq INTEGER PRIMARY KEY AUTOINCREMENT,kind TEXT NOT NULL,reference_id TEXT NOT NULL,created_at TEXT NOT NULL,payload TEXT NOT NULL);
CREATE INDEX events_kind_time ON events(kind,seq DESC);
CREATE INDEX events_reference ON events(reference_id);
CREATE TABLE benchmark_sessions(id TEXT PRIMARY KEY,protocol TEXT NOT NULL,address TEXT NOT NULL,created_at TEXT NOT NULL,result TEXT NOT NULL);
CREATE TABLE benchmark_measurements(session_id TEXT NOT NULL REFERENCES benchmark_sessions(id),sequence INTEGER NOT NULL,payload TEXT NOT NULL,PRIMARY KEY(session_id,sequence));
