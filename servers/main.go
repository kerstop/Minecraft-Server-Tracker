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

var Q *Queries

func Settup(db *sql.DB) {
	Q = New(db)
}

func ServePage() http.Handler {
	var mux = http.NewServeMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {

		ctx := context.Background()

		var serverList, err = Q.GetServers(ctx)
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

		ctx := context.Background()

		var url = r.FormValue("url")
		var port, err = strconv.Atoi(r.FormValue("port"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		_, err = Q.CreateServer(ctx, CreateServerParams{url, int64(port)})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			println("error:", err.Error())
			return
		}

		w.Header().Set("HX-Refresh", "true")
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("GET /{id}/edit", func(w http.ResponseWriter, r *http.Request) {

		ctx := context.Background()

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

		server, err := Q.GetServer(ctx, int64(id))
		if err != nil {
			fmt.Println("error:", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		t.ExecuteTemplate(w, "servers-tr-edit.html", server)

	})

	mux.HandleFunc("POST /{id}/edit", func(w http.ResponseWriter, r *http.Request) {

		ctx := context.Background()

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

		server, err := Q.GetServer(ctx, int64(id))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			println("Error:", err.Error())
			return
		}

		server.Url = url
		server.Port = int64(port)
		err = Q.UpdateServer(ctx, UpdateServerParams{
			Url:  server.Url,
			Port: server.Port,
			ID:   server.ID,
		})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			println("Error:", err.Error())
			return
		}

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

		ctx := context.Background()

		var id, err = strconv.Atoi(r.PathValue("id"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("invalid id provided"))
			return
		}

		server, err := Q.GetServer(ctx, int64(id))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			println("Error:", err.Error())
			return
		}

		Q.DeleteServer(ctx, int64(server.ID))

	})
	return mux
}

func (status ServerStatus) Save() error {
	ctx := context.Background()
	return Q.CreateServerStatus(ctx, CreateServerStatusParams{
		ServerID: status.ServerID,
		Time:     status.Time,
		Count:    status.Count,
	})
}

func (s Server) Ping(ctx context.Context) (ServerStatus, error) {

	started_at := time.Now()

	ctx, cancel := context.WithTimeout(ctx, time.Second*20)
	defer cancel()

	var players_online int64

	response, err := status.Modern(ctx, s.Url, uint16(s.Port))
	if err != nil {
		return ServerStatus{}, err
	} else {
		players_online = *response.Players.Online
	}

	fmt.Println(players_online, "players on", s.Url)

	return ServerStatus{
		ServerID: s.ID,
		Count:    players_online,
		Time:     started_at.Unix(),
	}, nil
}

type ServerList []Server

// get server stats and save them
func (servers ServerList) PingServers(ctx context.Context) {
	for _, server := range servers {
		stat, err := server.Ping(ctx)
		if err != nil {
			fmt.Println("Error:", err.Error())
			continue
		}
		err = stat.Save()

		if err != nil {
			fmt.Println("Error:", err.Error())
		}
	}
}
