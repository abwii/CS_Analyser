CREATE TABLE admin_users (
    id BIGSERIAL PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    role TEXT NOT NULL,
    password_hash TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO admin_users (email, role, password_hash) VALUES
    ('admin@example.com', 'admin', '$2a$10$PO.WVnpVbPFQ8oyVApdHc.j4drQ30c1vHxmfpKzIWG89vqkC3QgC2'),
    ('analyst@example.com', 'analyst', '$2a$10$PO.WVnpVbPFQ8oyVApdHc.j4drQ30c1vHxmfpKzIWG89vqkC3QgC2');
