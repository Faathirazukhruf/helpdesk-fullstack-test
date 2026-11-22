-- Down Migration: Remove Seed Users

DELETE FROM users 
WHERE email IN ('admin@example.com', 'agent@example.com', 'requester@example.com');
