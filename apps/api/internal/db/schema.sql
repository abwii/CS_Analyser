CREATE TABLE IF NOT EXISTS admin_users (
    id BIGSERIAL PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    role TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE admin_users ADD COLUMN IF NOT EXISTS password_hash TEXT;

CREATE TABLE IF NOT EXISTS competitive_maps (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO admin_users (email, role, password_hash) VALUES
    ('admin@example.com', 'admin', '$2a$10$PO.WVnpVbPFQ8oyVApdHc.j4drQ30c1vHxmfpKzIWG89vqkC3QgC2'),
    ('analyst@example.com', 'analyst', '$2a$10$PO.WVnpVbPFQ8oyVApdHc.j4drQ30c1vHxmfpKzIWG89vqkC3QgC2')
ON CONFLICT (email) DO UPDATE SET password_hash = EXCLUDED.password_hash;

INSERT INTO competitive_maps (name, active) VALUES
    ('de_dust2', true),
    ('de_mirage', true),
    ('de_inferno', true),
    ('de_nuke', true),
    ('de_ancient', true),
    ('de_anubis', true)
ON CONFLICT (name) DO NOTHING;
