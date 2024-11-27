DELETE FROM users_permissions
WHERE user_id = (SELECT id FROM users WHERE email = 'johnny@example.com')
  AND permission_id = (SELECT id FROM permissions WHERE code = 'readinglists:write');