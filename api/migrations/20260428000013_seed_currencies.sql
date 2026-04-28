-- +goose Up
-- Seed inicial de monedas: LATAM + USD + EUR.
-- ON CONFLICT DO NOTHING: idempotente — safe para re-runs y CI.
-- Monedas incluidas: COP, ARS, MXN, BRL, CLP, PEN, UYU, PYG, BOB, VES,
--                    CRC, GTQ, DOP, HNL, NIO, PAB, USD, EUR
INSERT INTO currencies (code, name, symbol) VALUES
  ('COP', 'Peso colombiano',      '$'),
  ('ARS', 'Peso argentino',       '$'),
  ('MXN', 'Peso mexicano',        '$'),
  ('BRL', 'Real brasileño',       'R$'),
  ('CLP', 'Peso chileno',         '$'),
  ('PEN', 'Sol peruano',          'S/'),
  ('UYU', 'Peso uruguayo',        '$U'),
  ('PYG', 'Guaraní paraguayo',    '₲'),
  ('BOB', 'Boliviano',            'Bs.'),
  ('VES', 'Bolívar venezolano',   'Bs.S'),
  ('CRC', 'Colón costarricense',  '₡'),
  ('GTQ', 'Quetzal guatemalteco', 'Q'),
  ('DOP', 'Peso dominicano',      'RD$'),
  ('HNL', 'Lempira hondureño',    'L'),
  ('NIO', 'Córdoba nicaragüense', 'C$'),
  ('PAB', 'Balboa panameño',      'B/.'),
  ('USD', 'Dólar estadounidense', '$'),
  ('EUR', 'Euro',                 '€')
ON CONFLICT (code) DO NOTHING;

-- +goose Down
DELETE FROM currencies
WHERE code IN (
  'COP', 'ARS', 'MXN', 'BRL', 'CLP', 'PEN', 'UYU', 'PYG',
  'BOB', 'VES', 'CRC', 'GTQ', 'DOP', 'HNL', 'NIO', 'PAB',
  'USD', 'EUR'
);
