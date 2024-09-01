package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/mcstatus-io/mcutil/v4/status"
)

type ServerStatus struct {
	PlayerCount int   `field:"count" json:"count"`
	Epoch       int64 `field:"time" json:"time"`
}

func PingServer(name string) ServerStatus {

	started_at := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	var players_online int64

	response, err := status.Modern(ctx, name, 25565)
	if err != nil {
		fmt.Printf("Error: [%s] at time %s\n", err.Error(), started_at.Format(time.Stamp))
		players_online = 0
	} else {
		players_online = *response.Players.Online
	}

	fmt.Println("Players Online:", players_online)

	return ServerStatus{
		PlayerCount: int(players_online),
		Epoch:       started_at.Unix(),
	}
}

func (status ServerStatus) Save(db *sql.DB) error {
	_, err := db.Exec(`
			INSERT INTO player_count
			VALUES (?, ?);
		`, status.Epoch, status.PlayerCount)
	return err
}

// period is how far into the past to look
func GetRecent(db *sql.DB, period time.Duration) ([]ServerStatus, error) {
	rows, err := db.Query(`
		SELECT * FROM player_count WHERE time > ? ORDER BY time ASC;
		`, time.Now().Add(period).Unix())
	if err != nil {
		return nil, err
	}

	var data []ServerStatus
	for rows.Next() {
		entry := ServerStatus{}
		rows.Scan(&entry.Epoch, &entry.PlayerCount)
		data = append(data, entry)
	}

	return data, nil

}
