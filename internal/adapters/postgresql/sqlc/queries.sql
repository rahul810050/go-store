-- name: ListProducts :many
SELECT * FROM products;


-- name: FindProductByID :one
SELECT * FROM products
WHERE id=$1;

-- name: CreateProduct :one
INSERT INTO products (name, price, quantity)
VALUES ($1, $2, $3)
RETURNING *;

-- name: CreateUser :one
INSERT INTO users (name, email, password_hash)
VALUES ($1, $2, $3)
RETURNING id, name, email, role, created_at, updated_at;

-- name: GetUserByEmail :one
SELECT *
FROM users
WHERE email = $1
LIMIT 1;

-- name: GetUserByID :one
SELECT id, name, email, role, created_at, updated_at
FROM users
WHERE id = $1
LIMIT 1;

-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (
    user_id,
    token_hash,
    family_id,
    expires_at
)
VALUES ($1, $2, $3, $4)
RETURNING *;


-- name: GetValidRefreshToken :one
SELECT *
FROM refresh_tokens
WHERE token_hash = $1
  AND expires_at > NOW()
  AND revoked_at IS NULL
LIMIT 1;


-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET revoked_at = NOW()
WHERE token_hash = $1
  AND revoked_at IS NULL;


-- name: RevokeRefreshTokenFamily :exec
UPDATE refresh_tokens
SET revoked_at = NOW()
WHERE family_id = $1
  AND revoked_at IS NULL;


-- name: CreateOrder :one
INSERT INTO orders (user_id, total_amount)
VALUES ($1, $2)
RETURNING *;


-- name: GetOrderByID :one
SELECT * FROM orders
WHERE id=$1 AND user_id=$2
LIMIT 1;


-- name: ListOrdersByUserID :many
SELECT * FROM orders
WHERE user_id=$1
ORDER BY created_at DESC;


-- name: CreateOrderItem :one
INSERT INTO order_items (
  order_id,
  product_id,
  quantity,
  unit_price,
  subtotal
)
VALUES ($1,$2,$3,$4, $5)
RETURNING *;


-- name: ListOrderItemsByOrderID :many
SELECT * FROM order_items
WHERE order_id=$1;


-- name: CreatePayment :one
INSERT INTO payments (order_id, amount)
VALUES ($1,$2)
RETURNING *;


-- name: GetPaymentByID :one
SELECT * FROM payments
WHERE id=$1
LIMIT 1;


-- name: ListPaymentsByOrderID :many
SELECT * FROM payments
WHERE order_id=$1
ORDER BY created_at DESC;


-- name: UpdatePaymentProviderOrderID :one
UPDATE payments
SET provider_order_id=$2,
    status='pending',
    updated_at=NOW()
WHERE id=$1
RETURNING *;


-- name: UpdatePaymentStatus :one
UPDATE payments
SET status=$2,
    provider_payment_id=$3,
    updated_at=NOW()
WHERE id=$1
RETURNING *;


-- name: GetProductForUpdate :one
SELECT *
FROM products
WHERE id = $1
FOR UPDATE;


-- name: DecreaseProductStock :one
UPDATE products
SET quantity=quantity-$2
WHERE id=$1 AND quantity >= $2
RETURNING *;