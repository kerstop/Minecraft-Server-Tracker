package servers

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/mcstatus-io/mcutil/v4/status"
)

var db *sql.DB

func Settup(_db *sql.DB) {
	db = _db
}

func ServePage() http.Handler {
	var mux = http.NewServeMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {

		var serverList, err = GetServers()
		if err != nil {
			println("Failed to read servers:", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		t, err := template.ParseGlob("servers/templates/servers*.html")
		if err != nil {
			fmt.Println("error:", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		err = t.ExecuteTemplate(w, "servers.html", serverList)
		if err != nil {
			fmt.Println("error:", err.Error())
		}

	})

	mux.HandleFunc("POST /", func(w http.ResponseWriter, r *http.Request) {

		var url = r.FormValue("url")
		var port, err = strconv.Atoi(r.FormValue("port"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		_, err = CreateServer(url, uint16(port))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			println("error:", err.Error())
			return
		}

		w.Header().Set("HX-Refresh", "true")
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("GET /{id}/edit", func(w http.ResponseWriter, r *http.Request) {
		var id, err = strconv.Atoi(r.PathValue("id"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("invalid id provided"))
			return
		}

		t, err := template.ParseGlob("servers/templates/servers*.html")
		if err != nil {
			fmt.Println("error:", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		server, err := GetServer(id)
		if err != nil {
			fmt.Println("error:", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		t.ExecuteTemplate(w, "servers-tr-edit.html", server)

	})

	mux.HandleFunc("POST /{id}/edit", func(w http.ResponseWriter, r *http.Request) {
		var id, err = strconv.Atoi(r.PathValue("id"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("invalid id provided"))
			return
		}

		var url = r.FormValue("url")
		port, err := strconv.Atoi(r.FormValue("port"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("bad port number"))
			return
		}

		server, err := GetServer(id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			println("Error:", err.Error())
			return
		}

		server.Url = url
		server.Port = uint16(port)
		server.Save()

		t, err := template.ParseGlob("servers/templates/servers*.html")
		if err != nil {
			fmt.Println("error:", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		t.ExecuteTemplate(w, "servers-tr.html", server)
	})

	mux.HandleFunc("DELETE /{id}", func(w http.ResponseWriter, r *http.Request) {
		var id, err = strconv.Atoi(r.PathValue("id"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("invalid id provided"))
			return
		}

		server, err := GetServer(id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			println("Error:", err.Error())
			return
		}

		server.Delete()

	})
	return mux
}

type ServerStatus struct {
	server_id   int   `field:"server_id"`
	PlayerCount int   `field:"count" json:"count"`
	Epoch       int64 `field:"time" json:"time"`
}

func (status ServerStatus) Save() error {
	_, err := db.Exec(`
			INSERT INTO player_count
			VALUES (?, ?, ?);
		`, status.server_id, status.Epoch, status.PlayerCount)
	return err
}

type Server struct {
	Id   int
	Url  string
	Port uint16
}

func GetServer(id int) (Server, error) {
	row := db.QueryRow(`SELECT * FROM servers WHERE id = ?;`, id)
	server := Server{}
	err := row.Scan(&server.Id, &server.Url, &server.Port)
	if err != nil {
		return Server{}, err
	}

	return server, nil
}

func CreateServer(url string, port uint16) (Server, error) {
	row := db.QueryRow(`INSERT INTO servers (url, port) VALUES (?, ?) RETURNING *;`, url, port)
	server := Server{}
	err := row.Scan(&server.Id, &server.Url, &server.Port)
	if err != nil {
		return Server{}, err
	}
	return server, nil
}

func (s Server) Save() error {
	_, err := db.Exec(`UPDATE servers SET url = ?, port = ? WHERE id = ?;`, s.Url, s.Port, s.Id)
	return err
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
		server_id:   s.Id,
		PlayerCount: int(players_online),
		Epoch:       started_at.Unix(),
	}
}

func (s Server) Delete() error {
	_, err := db.Exec(`DELETE FROM servers WHERE id = ?;`, s.Id)
	return err
}

// period is how far into the past to look
func (server Server) GetRecent(period time.Duration) ([]ServerStatus, error) {
	rows, err := db.Query(`
		SELECT server_id, count, time FROM player_count WHERE time > ? and server_id = ? ORDER BY time ASC;
		`, time.Now().Add(period).Unix(), server.Id)
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

type ServerList []Server

func GetServers() (ServerList, error) {
	rows, err := db.Query(`SELECT * FROM servers;`)
	if err != nil {
		return nil, err
	}
	var servers []Server
	for rows.Next() {
		var server Server
		err = rows.Scan(&server.Id, &server.Url, &server.Port)
		if err != nil {
			println("Error:", err.Error())
		}
		servers = append(servers, server)

	}

	return ServerList(servers), nil
}

// get server stats and save them
func (servers ServerList) PingServers() {
	var err error
	for _, server := range servers {
		err = server.Ping().Save()
		if err != nil {
			fmt.Println("Error:", err.Error())
		}
	}
}
