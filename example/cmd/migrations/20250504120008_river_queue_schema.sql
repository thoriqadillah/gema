-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS message_queue;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP SCHEMA IF EXISTS message_queue CASCADE;
-- +goose StatementEnd
