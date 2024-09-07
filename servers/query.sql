-- name: GetServers :many
SELECT * FROM servers;

-- name: GetServer :one
SELECT * FROM servers WHERE id = ?;

-- name: CreateServer :one
INSERT INTO servers (url, port) VALUES (?, ?) RETURNING *;

-- name: UpdateServer :exec
UPDATE servers SET url = ?, port = ? WHERE id = ?;

-- name: DeleteServer :exec
DELETE FROM servers WHERE id = ?;

-- name: GetServerStatus :many
SELECT count, time FROM server_status WHERE time > @unix_start_time and server_id = ? ORDER BY time ASC;

-- name: CreateServerStatus :exec
INSERT INTO server_status (server_id, time, count) VALUES (?, ?, ?);
