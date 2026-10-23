ALTER TABLE ytv_source_settings
    ADD COLUMN tags JSONB,
    ADD COLUMN recommended BOOLEAN NOT NULL DEFAULT FALSE,
    ADD CONSTRAINT ytv_source_tags_array CHECK (tags IS NULL OR jsonb_typeof(tags) = 'array');
