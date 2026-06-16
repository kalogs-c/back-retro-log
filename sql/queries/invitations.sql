-- name: CreateInvitation :one
INSERT INTO invitations (token, created_by, expires_at)
VALUES (?, ?, ?)
RETURNING id, token, created_by, expires_at, used_by, used_at, created_at;

-- name: GetInvitationByToken :one
SELECT id, token, created_by, expires_at, used_by, used_at, created_at
FROM invitations
WHERE token = ?;

-- name: UseInvitation :exec
UPDATE invitations
SET used_by = ?, used_at = datetime('now')
WHERE id = ? AND used_by IS NULL;

-- name: ListInvitationsByUser :many
SELECT id, token, created_by, expires_at, used_by, used_at, created_at
FROM invitations
WHERE created_by = ?
ORDER BY created_at DESC;
