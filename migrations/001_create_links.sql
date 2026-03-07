CREATE TABLE links
(
    id         uuid PRIMARY KEY,
    short_code TEXT UNIQUE,
    long_url   TEXT,
    disabled   BOOLEAN,
    created_at TIMESTAMP
);