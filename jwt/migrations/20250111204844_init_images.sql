-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS images (
    id SERIAL PRIMARY KEY,
    path VARCHAR(150) NOT NULL,
    path_id INTEGER NOT NULL,
    creator_id UUID REFERENCES users(id) ON DELETE SET NULL,
    creation_time timestamp default current_timestamp NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE images;
-- +goose StatementEnd
