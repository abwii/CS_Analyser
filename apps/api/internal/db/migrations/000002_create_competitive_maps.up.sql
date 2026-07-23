CREATE TABLE competitive_maps (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO competitive_maps (name, active) VALUES
    ('de_dust2', true),
    ('de_mirage', true),
    ('de_inferno', true),
    ('de_nuke', true),
    ('de_ancient', true),
    ('de_anubis', true);
