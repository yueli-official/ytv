CREATE TABLE ytv_site_profiles (
    site_key TEXT PRIMARY KEY,
    name TEXT NOT NULL CHECK (name <> '' AND length(name) <= 80),
    description TEXT NOT NULL CHECK (length(description) <= 300),
    revision BIGINT NOT NULL DEFAULT 1 CHECK (revision >= 1),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
