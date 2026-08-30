ALTER TABLE ytv_source_settings
    ADD COLUMN display_name TEXT NOT NULL DEFAULT '',
    ADD COLUMN installed BOOLEAN NOT NULL DEFAULT TRUE;
