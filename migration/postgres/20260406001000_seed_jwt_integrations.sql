-- +goose Up
-- +goose StatementBegin
INSERT INTO integrations (
    id, name, description, base_url, is_active,
    jwt_access_ttl_seconds, jwt_refresh_ttl_seconds,
    jwt_binding_mode, jwt_max_sessions,
    created_at, updated_at
) VALUES (
    gen_random_uuid(), 'gct-admin', 'JWT integration: gct-admin', 'internal', true,
    300, 28800,
    'strict', 2,
    NOW(), NOW()
)
ON CONFLICT (name) DO NOTHING;
-- +goose StatementEnd

-- +goose StatementBegin
INSERT INTO integrations (
    id, name, description, base_url, is_active,
    jwt_access_ttl_seconds, jwt_refresh_ttl_seconds,
    jwt_binding_mode, jwt_max_sessions,
    created_at, updated_at
) VALUES (
    gen_random_uuid(), 'gct-client', 'JWT integration: gct-client', 'internal', true,
    900, 2592000,
    'warn', 5,
    NOW(), NOW()
)
ON CONFLICT (name) DO NOTHING;
-- +goose StatementEnd

-- +goose StatementBegin
INSERT INTO integrations (
    id, name, description, base_url, is_active,
    jwt_access_ttl_seconds, jwt_refresh_ttl_seconds,
    jwt_binding_mode, jwt_max_sessions,
    created_at, updated_at
) VALUES (
    gen_random_uuid(), 'gct-mobile', 'JWT integration: gct-mobile', 'internal', true,
    1800, 7776000,
    'warn', 10,
    NOW(), NOW()
)
ON CONFLICT (name) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM integrations WHERE name IN ('gct-admin', 'gct-client', 'gct-mobile');
-- +goose StatementEnd
