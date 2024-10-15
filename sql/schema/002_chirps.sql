-- +goose Up
CREATE TABLE chirps(
  id UUID DEFAULT gen_random_uuid(),
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  body TEXT NOT NULL,
  user_id UUID not NULL,
  PRIMARY KEY (id),
  CONSTRAINT fk_user
  FOREIGN KEY (user_id)
  REFERENCES users(id)
  ON DELETE CASCADE
);

-- +goose Down
drop TABLE chirps;