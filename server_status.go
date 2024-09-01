package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/mcstatus-io/mcutil/v4/status"
)

type ServerStatus struct {
	server_id   int   `field:"server_id"`
	PlayerCount int   `field:"count" json:"count"`
	Epoch       int64 `field:"time" json:"time"`
}

type Server struct {
	id   int
	Url  string
	Port uint16
}

func (status ServerStatus) Save(db *sql.DB) error {
	_, err := db.Exec(`
			INSERT INTO player_count
			VALUES (?, ?, ?);
		`, status.server_id, status.Epoch, status.PlayerCount)
	return err
}

func GetServers(db *sql.DB) Servers {
	rows, err := db.Query(`SELECT * FROM servers;`)
	if err != nil {
		println("Error:", err.Error())
	}
	var servers []Server
	for rows.Next() {
		var server Server
		err = rows.Scan(&server.id, &server.Url, &server.Port)
		if err != nil {
			println("Error:", err.Error())
		}
		servers = append(servers, server)

	}

	return Servers(servers)
}

type Servers []Server

// get server stats and save them
func (servers Servers) PingServers() {
	var err error
	for _, server := range servers {
		err = server.Ping().Save(db)
		if err != nil {
			fmt.Println("Error:", err.Error())
		}
	}
}

func (s Server) Ping() ServerStatus {

	started_at := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	var players_online int64

	response, err := status.Modern(ctx, s.Url, s.Port)
	if err != nil {
		fmt.Printf("Error: [%s] at time %s\n", err.Error(), started_at.Format(time.Stamp))
		players_online = 0
	} else {
		players_online = *response.Players.Online
	}

	fmt.Println(players_online, "players on", s.Url)

	return ServerStatus{
		server_id:   s.id,
		PlayerCount: int(players_online),
		Epoch:       started_at.Unix(),
	}
}

// period is how far into the past to look
func (server Server) GetRecent(db *sql.DB, period time.Duration) ([]ServerStatus, error) {
	rows, err := db.Query(`
		SELECT server_id, count, time FROM player_count WHERE time > ? and server_id = ? ORDER BY time ASC;
		`, time.Now().Add(period).Unix(), server.id)
	if err != nil {
		return nil, err
	}

	var data []ServerStatus
	for rows.Next() {
		entry := ServerStatus{}
		rows.Scan(&entry.server_id, &entry.PlayerCount, &entry.Epoch)
		data = append(data, entry)
	}

	return data, nil

}
