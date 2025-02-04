// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: query.sql

package servers

import (
	"context"
)

const createServer = `-- name: CreateServer :one
INSERT INTO servers (url, port) VALUES (?, ?) RETURNING id, url, port
`

type CreateServerParams struct {
	Url  string `json:"url"`
	Port int64  `json:"port"`
}

func (q *Queries) CreateServer(ctx context.Context, arg CreateServerParams) (Server, error) {
	row := q.db.QueryRowContext(ctx, createServer, arg.Url, arg.Port)
	var i Server
	err := row.Scan(&i.ID, &i.Url, &i.Port)
	return i, err
}

const createServerStatus = `-- name: CreateServerStatus :exec
INSERT INTO server_status (server_id, time, count) VALUES (?, ?, ?)
`

type CreateServerStatusParams struct {
	ServerID int64 `json:"server_id"`
	Time     int64 `json:"time"`
	Count    int64 `json:"count"`
}

func (q *Queries) CreateServerStatus(ctx context.Context, arg CreateServerStatusParams) error {
	_, err := q.db.ExecContext(ctx, createServerStatus, arg.ServerID, arg.Time, arg.Count)
	return err
}

const deleteServer = `-- name: DeleteServer :exec
DELETE FROM servers WHERE id = ?
`

func (q *Queries) DeleteServer(ctx context.Context, id int64) error {
	_, err := q.db.ExecContext(ctx, deleteServer, id)
	return err
}

const getServer = `-- name: GetServer :one
SELECT id, url, port FROM servers WHERE id = ?
`

func (q *Queries) GetServer(ctx context.Context, id int64) (Server, error) {
	row := q.db.QueryRowContext(ctx, getServer, id)
	var i Server
	err := row.Scan(&i.ID, &i.Url, &i.Port)
	return i, err
}

const getServerStatus = `-- name: GetServerStatus :many
SELECT count, time FROM server_status WHERE time > ? and server_id = ? ORDER BY time ASC
`

type GetServerStatusParams struct {
	UnixStartTime int64 `json:"unix_start_time"`
	ServerID      int64 `json:"server_id"`
}

type GetServerStatusRow struct {
	Count int64 `json:"count"`
	Time  int64 `json:"time"`
}

func (q *Queries) GetServerStatus(ctx context.Context, arg GetServerStatusParams) ([]GetServerStatusRow, error) {
	rows, err := q.db.QueryContext(ctx, getServerStatus, arg.UnixStartTime, arg.ServerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetServerStatusRow
	for rows.Next() {
		var i GetServerStatusRow
		if err := rows.Scan(&i.Count, &i.Time); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getServers = `-- name: GetServers :many
SELECT id, url, port FROM servers
`

func (q *Queries) GetServers(ctx context.Context) ([]Server, error) {
	rows, err := q.db.QueryContext(ctx, getServers)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Server
	for rows.Next() {
		var i Server
		if err := rows.Scan(&i.ID, &i.Url, &i.Port); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateServer = `-- name: UpdateServer :exec
UPDATE servers SET url = ?, port = ? WHERE id = ?
`

type UpdateServerParams struct {
	Url  string `json:"url"`
	Port int64  `json:"port"`
	ID   int64  `json:"id"`
}

func (q *Queries) UpdateServer(ctx context.Context, arg UpdateServerParams) error {
	_, err := q.db.ExecContext(ctx, updateServer, arg.Url, arg.Port, arg.ID)
	return err
}
