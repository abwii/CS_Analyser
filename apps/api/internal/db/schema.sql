CREATE TABLE IF NOT EXISTS admin_users (
    id BIGSERIAL PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    role TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS competitive_maps (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO admin_users (email, role) VALUES
    ('admin@example.com', 'admin'),
    ('analyst@example.com', 'analyst')
ON CONFLICT (email) DO NOTHING;

INSERT INTO competitive_maps (name, active) VALUES
    ('de_dust2', true),
    ('de_mirage', true),
    ('de_inferno', true),
    ('de_nuke', true),
    ('de_ancient', true),
    ('de_anubis', true)
ON CONFLICT (name) DO NOTHING;
