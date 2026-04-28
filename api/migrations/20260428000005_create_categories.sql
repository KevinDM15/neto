-- +goose Up
-- Categorías jerárquicas de un nivel (parent_id auto-referencial).
-- ON DELETE SET NULL: eliminar categoría padre no borra las hijas, las deja como raíz.
CREATE TABLE categories (
  id         UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id    UUID         NOT NULL,
  name       VARCHAR(100) NOT NULL,
  parent_id  UUID         REFERENCES categories(id) ON DELETE SET NULL,
  icon       VARCHAR(50),
  created_at TIMESTAMPTZ  NOT NULL DEFAULT now()
);
CREATE INDEX idx_categories_user_id   ON categories(user_id);
CREATE INDEX idx_categories_parent_id ON categories(parent_id);

-- +goose Down
DROP TABLE IF EXISTS categories;
