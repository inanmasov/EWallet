DO $$ 
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM pg_database WHERE datname = 'db_infotecs') THEN 
        CREATE DATABASE db_infotecs; 
    END IF; 
END $$;

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