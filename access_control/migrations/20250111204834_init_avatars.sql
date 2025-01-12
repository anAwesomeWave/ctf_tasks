-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS avatars (
    id SERIAL PRIMARY KEY,
    path VARCHAR(150) NOT NULL,
    owner_id UUID REFERENCES users(id) ON DELETE SET NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE avatars;
-- +goose StatementEnd
