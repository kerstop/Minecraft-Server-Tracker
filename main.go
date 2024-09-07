package main

import (
	"blockgame_ping/servers"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB
var templates *template.Template

func main() {

	var err error
	db, err = sql.Open("sqlite3", "file:data.sqlite")
	if err != nil {
		fmt.Printf("problem opening file: %s\n", err.Error())
		os.Exit(1)
	}
	servers.Settup(db)

	_, err = db.Exec(`
			CREATE TABLE IF NOT EXISTS servers (
				id INTEGER PRIMARY KEY,
				url TEXT UNIQUE NOT NULL,
				port INTEGER DEFAULT 25565 NOT NULL
			);

			CREATE TABLE IF NOT EXISTS player_count (
				server_id INTEGER,
				time INTEGER,
				count INTEGER NOT NULL,
				PRIMARY KEY (server_id, time),
				FOREIGN KEY (server_id) REFERENCES servers(id) ON DELETE CASCADE
			);
		`)
	if err != nil {
		fmt.Printf("unable to create table: %s\n", err.Error())
		os.Exit(1)
	}
	http.Handle("/servers/", http.StripPrefix("/servers", servers.ServePage()))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.HandleFunc("/static/bundle.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "frontend/dist/bundle.js")
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		ctx := context.Background()

		type templateData []struct {
			Title string
			Data  string
		}
		var data templateData

		serverList, err := servers.Q.GetServers(ctx)
		if err != nil {
			println("Failed to read servers:", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		for _, v := range serverList {

			server_data, err := servers.Q.GetServerStatus(ctx, servers.GetServerStatusParams{
				UnixStartTime: time.Now().Add(time.Hour * -5).Unix(),
				ServerID:      v.ID,
			})
			if err != nil {
				fmt.Println(err)
				return
			}

			server_data_json, err := json.Marshal(server_data)

			data = append(data, struct {
				Title string
				Data  string
			}{
				v.Url, base64.StdEncoding.EncodeToString(server_data_json),
			})

		}
		w.WriteHeader(http.StatusOK)
		t, err := template.ParseGlob("templates/*.html")
		if err != nil {
			fmt.Println("Error:", err.Error())
		}
		t.ExecuteTemplate(w, "index.html", data)
	})

	go func() {
		for {
			go func() {
				ctx := context.Background()
				s, err := servers.Q.GetServers(ctx)
				if err != nil {
					println("Failed to read servers:", err.Error())
					return
				}

				servers.ServerList(s).PingServers(ctx)
			}()
			time.Sleep(time.Minute)
		}
	}()

	fmt.Println("Starting HTTP Server")
	http.ListenAndServe("localhost:8080", nil)
}
