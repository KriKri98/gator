-- name: CreateFeed :exec
INSERT INTO feeds(id, created_at, updated_at, name, url, user_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
);

-- name: GetFeeds :many
SELECT * FROM feeds;

-- name: GetFeedsURL :one
SELECT * FROM feeds WHERE feeds.url = $1;

-- name: MarkFeedFetched :exec
UPDATE feeds SET updated_at = $1, last_fetched_at = $1 WHERE id = $2;

-- name: GetNextFeedToFetch :one
SELECT * FROM feeds WHERE user_id = $1 ORDER BY last_fetched_at ASC NULLS FIRST LIMIT 1;

