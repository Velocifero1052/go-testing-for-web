package main

import (
	"encoding/gob"
	"flag"
	"github.com/alexedwards/scs/v2"
	"log"
	"net/http"
	"testingCourserWeb/pkg/data"
	"testingCourserWeb/pkg/repository"
	"testingCourserWeb/pkg/repository/dbrepo"
)

type application struct {
	DSN     string
	DB      repository.DatabaseRepo
	Session *scs.SessionManager
}

func main() {
	gob.Register(data.User{})
	// set up an app config
	app := application{}

	flag.StringVar(&app.DSN, "dsn", "host=localhost port=5433 user=postgres password=postgres dbname=users sslmode=disable timezone=UTC connect_timeout=5", "Postgres connection")
	flag.Parse()

	conn, err := app.connectToDB()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	app.DB = &dbrepo.PostgresDBRepo{DB: conn}
	//get a session manager
	app.Session = getSession()

	// get application routes
	mux := app.routes()

	// print out a message
	log.Println("Starting server on port 8080...")

	// start the server
	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}
