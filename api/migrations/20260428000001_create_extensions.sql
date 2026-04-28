-- +goose Up
-- Habilitar uuid-ossp para uuid_generate_v4() usado en todos los PKs
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- +goose Down
DROP EXTENSION IF EXISTS "uuid-ossp";
