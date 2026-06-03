ALTER TABLE users ADD COLUMN password TEXT NOT NULL DEFAULT '123';

UPDATE users
SET player_name = 'Blazor',
    role_id = 3,
    password = '123',
    api_token = 'blazor-admin-token'
WHERE id = 1;

UPDATE users
SET password = 'disabled-' || id,
    api_token = 'disabled-token-' || id
WHERE id <> 1;
