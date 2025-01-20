-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    login VARCHAR(100) NOT NULL UNIQUE,
    password_hash VARCHAR(100) NOT NULL,  -- bcrypt hash length - 60 chars.
    is_admin BOOLEAN NOT NULL DEFAULT FALSE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE users;
-- +goose StatementEnd
