package main

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB
var debug bool
var templates *template.Template

func main() {

	flag.BoolVar(&debug, "debug", false, "enable dynamic loading of html templates")
	flag.Parse()
	if debug {
		fmt.Println("Running in debug mode")
	}

	var err error
	db, err = sql.Open("sqlite3", "file:data.sqlite")
	if err != nil {
		fmt.Printf("problem opening file: %s\n", err.Error())
		os.Exit(1)
	}

	_, err = db.Exec(`
			CREATE TABLE IF NOT EXISTS player_count (
				time INTEGER PRIMARY KEY,
				count INTEGER
			);
		`)
	if err != nil {
		fmt.Printf("unable to create table: %s\n", err.Error())
		os.Exit(1)
	}
	http.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.HandleFunc("GET /static/bundle.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "frontend/dist/bundle.js")
	})

	http.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {

		data, err := GetRecent(db, time.Hour*-5)
		if err != nil {
			fmt.Println(err)
			return
		}

		data_json, err := json.Marshal(data)

		w.WriteHeader(http.StatusOK)
		getTemplate("index.html").Execute(w, struct {
			Title string
			Data  string
		}{
			"blockgame", base64.StdEncoding.EncodeToString(data_json),
		})
	})

	go func() {
		for {
			go func() { PingServer("mc.blockgame.info").Save(db) }()
			time.Sleep(time.Minute)
		}
	}()

	fmt.Println("Starting HTTP Server")
	http.ListenAndServe("localhost:8080", nil)
}

func getTemplate(name string) *template.Template {
	if debug {
		t, err := template.ParseGlob("templates/*.html")
		if err != nil {
			fmt.Println(err.Error())
			return nil
		}
		return t.Lookup(name)
	}

	if templates == nil {

		t, err := template.ParseGlob("templates/*.html")
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		templates = t
	}

	return templates.Lookup(name)
}
