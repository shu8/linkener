CREATE TABLE users (
    username TEXT PRIMARY KEY,
    password TEXT
);

CREATE TABLE access_tokens (
    username TEXT PRIMARY KEY,
    access_token TEXT,
    expiry DATETIME DEFAULT (datetime('now', '+1 hour'))
);

.quit
