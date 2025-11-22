CREATE TABLE IF NOT EXISTS metadata (
    id VARCHAR PRIMARY KEY,
    name TEXT,
    location TEXT,
    capacity INT
);

CREATE TABLE IF NOT EXISTS reservations (
    id SERIAL PRIMARY KEY,
    record_id TEXT,
    record_type TEXT,
    user_id TEXT,
    start_time TIMESTAMP,
    end_time TIMESTAMP,
    status TEXT
);
