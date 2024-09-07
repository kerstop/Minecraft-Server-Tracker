package main

import (
	"blockgame_ping/servers"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB
var templates *template.Template
var connections map[uint64]*websocket.Conn = make(map[uint64]*websocket.Conn)
var connectionsCounter uint64 = 0

func main() {

	var err error
	db, err = sql.Open("sqlite3", "file:data.sqlite")
	if err != nil {
		fmt.Printf("problem opening file: %s\n", err.Error())
		os.Exit(1)
	}
	servers.Settup(db)

	http.Handle("/servers/", http.StripPrefix("/servers", servers.ServePage()))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.HandleFunc("/static/bundle.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "frontend/dist/bundle.js")
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		ctx := context.Background()

		type templateData struct {
			Title string
			Data  string
			ID    int64
		}
		var data []templateData

		var hours int
		if r.URL.Query().Has("history") {
			hours, err = strconv.Atoi(r.URL.Query().Get("history"))
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		} else {
			hours = 5
		}

		serverList, err := servers.Q.GetServers(ctx)
		if err != nil {
			println("Failed to read servers:", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		for _, server := range serverList {

			server_data, err := servers.Q.GetServerStatus(ctx, servers.GetServerStatusParams{
				UnixStartTime: time.Now().Add(time.Hour * time.Duration(-hours)).Unix(),
				ServerID:      server.ID,
			})
			if err != nil {
				fmt.Println(err)
				return
			}

			server_data_json, err := json.Marshal(server_data)

			data = append(data, templateData{
				server.Url, base64.StdEncoding.EncodeToString(server_data_json), server.ID,
			})

		}
		w.WriteHeader(http.StatusOK)
		t, err := template.ParseGlob("templates/*.html")
		if err != nil {
			fmt.Println("Error:", err.Error())
		}
		t.ExecuteTemplate(w, "index.html", data)
	})

	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, http.Header{})
		if err != nil {
			log.Println(err)
		}
		connections[connectionsCounter] = conn
		connectionsCounter += 1
	})

	go func() {
		for {
			go runSensor()
			time.Sleep(time.Minute)
		}
	}()

	fmt.Println("Starting HTTP Server")
	http.ListenAndServe("localhost:8080", nil)
}

func runSensor() {

	ctx := context.Background()
	serverList, err := servers.Q.GetServers(ctx)
	if err != nil {
		println("Failed to read servers:", err.Error())
		return
	}

	var updates = servers.ServerList(serverList).PingServers(ctx)

	for conID, connection := range connections {
		err = connection.WriteJSON(updates)
		if err != nil {
			delete(connections, conID)
		}
	}
	for _, update := range updates {
		server, _ := servers.Q.GetServer(ctx, update.ServerID)
		println(update.Count, "players on", server.Url)
	}
	println(len(connections), "connections recieving updates")

}
