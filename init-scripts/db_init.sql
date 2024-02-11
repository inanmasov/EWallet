CREATE TABLE IF NOT EXISTS Wallet (
  id SERIAL,
  balance NUMERIC NOT NULL
);

CREATE TABLE IF NOT EXISTS History (
  id INT NOT NULL,
  time TIMESTAMP NOT NULL,
  from_id INT NOT NULL,
  to_id INT NOT NULL,
  amount NUMERIC NOT NULL
);