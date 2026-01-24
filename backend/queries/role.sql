-- name: CreateRole :one
INSERT INTO roles (
    id,
    club_id,
    name,
    pays_base_fee,
    created_at,
    updated_at
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
RETURNING *;

-- name: GetRoleByID :one
SELECT * FROM roles
WHERE id = $1;

-- name: GetRolesByClubID :many
SELECT * FROM roles
WHERE club_id = $1
ORDER BY created_at DESC;

-- name: UpdateRole :one
UPDATE roles
SET
    name = $1,
    pays_base_fee = $2,
    updated_at = $3
WHERE id = $4
RETURNING *;

-- name: DeleteRole :exec
DELETE FROM roles
WHERE id = $1;

-- name: AddRolePermission :one
INSERT INTO role_permissions (
    id,
    role_id,
    entity_type,
    permission_type,
    created_at
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING *;

-- name: RemoveRolePermission :exec
DELETE FROM role_permissions
WHERE role_id = $1 AND entity_type = $2 AND permission_type = $3;

-- name: GetRolePermissions :many
SELECT * FROM role_permissions
WHERE role_id = $1
ORDER BY entity_type, permission_type;

-- name: HasRolePermission :one
SELECT COUNT(*) > 0
FROM role_permissions
WHERE role_id = $1 AND entity_type = $2 AND permission_type = $3;

