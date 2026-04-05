-- +goose Up
-- +goose StatementBegin
INSERT INTO users (id, role_id, username, email, phone, password_hash, active, is_approved)
SELECT 
    '00000000-0000-0000-0000-000000000001', 
    id, 
    'admin', 
    'admin@test.com', 
    '+998901234567', 
    '$2a$10$vI8aWBnW3fID.97.kHjSLe9M8U.RE9C7kY1R.9WJ.H.9WJ.H.9WJ.H', 
    true, 
    true
FROM role 
WHERE name = 'admin'
ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM users WHERE id = '00000000-0000-0000-0000-000000000001';
-- +goose StatementEnd
