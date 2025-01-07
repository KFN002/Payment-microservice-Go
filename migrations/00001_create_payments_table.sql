-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE payments (
	id uuid PRIMARY KEY DEFAULT uuid_generate_v4 (),
	from_user_id uuid NOT NULL,
	to_user_id uuid NOT NULL,
	amount double precision NOT NULL,
	currency varchar(3) NOT NULL,
	status varchar(20) NOT NULL DEFAULT 'PENDING',
	created_at timestamptz NOT NULL DEFAULT NOW(),
	updated_at timestamptz NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS payments;

-- DROP TABLE IF EXISTS users;
