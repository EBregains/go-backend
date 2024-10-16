-- +goose Up
Alter table users
Add column is_chirpy_red BOOLEAN DEFAULT(FALSE) NOT NULL;

-- +goose Down
Alter table users
Drop column is_chirpy_red;