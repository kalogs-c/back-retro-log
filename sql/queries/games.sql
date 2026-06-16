-- name: CreateGame :one
INSERT INTO games (rawg_id, title, cover_url, description, release_date)
VALUES (?, ?, ?, ?, ?)
RETURNING id, rawg_id, title, cover_url, description, release_date, created_at;

-- name: GetGameByRawgID :one
SELECT id, rawg_id, title, cover_url, description, release_date, created_at
FROM games
WHERE rawg_id = ?;

-- name: GetGameByID :one
SELECT id, rawg_id, title, cover_url, description, release_date, created_at
FROM games
WHERE id = ?;

-- name: SearchGames :many
SELECT id, rawg_id, title, cover_url, description, release_date, created_at
FROM games
WHERE title LIKE '%' || ? || '%'
ORDER BY title;
