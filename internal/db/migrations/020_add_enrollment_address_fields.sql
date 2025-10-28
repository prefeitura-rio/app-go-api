-- +goose Up
-- +goose StatementBegin
ALTER TABLE inscricoes ADD COLUMN age INTEGER;
ALTER TABLE inscricoes ADD COLUMN address VARCHAR(500);
ALTER TABLE inscricoes ADD COLUMN neighborhood VARCHAR(100);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE inscricoes DROP COLUMN neighborhood;
ALTER TABLE inscricoes DROP COLUMN address;
ALTER TABLE inscricoes DROP COLUMN age;
-- +goose StatementEnd
