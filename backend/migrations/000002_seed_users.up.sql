-- Up Migration: Seed Initial Users for Each Role
-- Password untuk semua seed user: Password123!
-- Bcrypt Hash (cost 10): $2a$10$/Cybguq/EsrAT48rjh58P.t2FxpE/BZdMDgkbJPeXqca/e9cvz6VO

INSERT INTO users (name, email, password_hash, role)
VALUES 
    ('System Admin', 'admin@example.com', '$2a$10$/Cybguq/EsrAT48rjh58P.t2FxpE/BZdMDgkbJPeXqca/e9cvz6VO', 'admin'),
    ('Support Agent', 'agent@example.com', '$2a$10$/Cybguq/EsrAT48rjh58P.t2FxpE/BZdMDgkbJPeXqca/e9cvz6VO', 'agent'),
    ('Requester User', 'requester@example.com', '$2a$10$/Cybguq/EsrAT48rjh58P.t2FxpE/BZdMDgkbJPeXqca/e9cvz6VO', 'requester')
ON CONFLICT (email) DO NOTHING;
