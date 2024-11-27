INSERT INTO users_permissions (user_id, permission_id)
VALUES (
    (SELECT id FROM users WHERE email = 'johnny@example.com'),
    (SELECT id FROM permissions WHERE code = 'comments:write')
);
