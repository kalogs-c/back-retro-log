-- name: GetCatalogEntryByUserAndGame :one
SELECT id FROM catalog_entries
WHERE user_id = ? AND game_id = ?;

-- name: CreateCatalogEntry :one
INSERT INTO catalog_entries (user_id, game_id, status)
VALUES (?, ?, ?)
RETURNING id, user_id, game_id, status, created_at, updated_at;

-- name: GetCatalogEntry :one
SELECT ce.id, ce.user_id, ce.game_id, ce.status, ce.created_at, ce.updated_at,
       g.title, g.cover_url, g.description, g.release_date
FROM catalog_entries ce
JOIN games g ON ce.game_id = g.id
WHERE ce.id = ? AND ce.user_id = ?;

-- name: ListCatalogEntries :many
SELECT ce.id, ce.user_id, ce.game_id, ce.status, ce.created_at, ce.updated_at,
       g.title, g.cover_url, g.description, g.release_date
FROM catalog_entries ce
JOIN games g ON ce.game_id = g.id
WHERE ce.user_id = ?
ORDER BY ce.updated_at DESC;

-- name: ListCatalogEntriesByStatus :many
SELECT ce.id, ce.user_id, ce.game_id, ce.status, ce.created_at, ce.updated_at,
       g.title, g.cover_url, g.description, g.release_date
FROM catalog_entries ce
JOIN games g ON ce.game_id = g.id
WHERE ce.user_id = ? AND ce.status = ?
ORDER BY ce.updated_at DESC;

-- name: UpdateCatalogEntryStatus :exec
UPDATE catalog_entries
SET status = ?, updated_at = datetime('now')
WHERE id = ? AND user_id = ?;

-- name: DeleteCatalogEntry :exec
DELETE FROM catalog_entries WHERE id = ? AND user_id = ?;
