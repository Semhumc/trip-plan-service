-- +goose Up
-- +goose StatementBegin
ALTER TABLE trips ADD COLUMN start_position TEXT;
ALTER TABLE trips ADD COLUMN finish_position TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE trips DROP COLUMN start_position;
ALTER TABLE trips DROP COLUMN finish_position;
-- +goose StatementEnd
