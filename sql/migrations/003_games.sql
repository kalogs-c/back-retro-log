CREATE TABLE IF NOT EXISTS games (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    rawg_id INTEGER UNIQUE,
    title TEXT NOT NULL,
    cover_url TEXT,
    description TEXT,
    release_date TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
