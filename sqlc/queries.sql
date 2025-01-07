-- name: CreatePayment :one
INSERT INTO payments (id, from_user_id, to_user_id, amount, currency, status)
	VALUES ($1, $2, $3, $4, $5, 'PENDING')
RETURNING
	id;

-- name: GetPaymentByID :one
SELECT
	*
FROM
	payments
WHERE
	id = $1;

-- name: RefundPayment :exec
UPDATE
	payments
SET
	status = 'REFUNDED'
WHERE
	id = $1;

-- name: GetPaymentHistory :many
SELECT
	*
FROM
	payments
WHERE
	from_user_id = $1
LIMIT $2 OFFSET $3;

-- name: UpdatePaymentStatus :exec
UPDATE
	payments
SET
	status = $1
WHERE
	id = $2;
