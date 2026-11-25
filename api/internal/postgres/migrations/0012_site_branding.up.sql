ALTER TABLE ytv_site_profiles
    ADD COLUMN logo_url text NOT NULL DEFAULT '',
    ADD COLUMN favicon_url text NOT NULL DEFAULT '';
