INSERT INTO users_permissions
SELECT id, (SELECT id FROM permissions WHERE code = 'comments:read') 
FROM users;
