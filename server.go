package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/coreos/go-oidc"
	_ "github.com/lib/pq"
)

func main() {
	var deps Dependencies

	deps.log = log.New(os.Stdout, "", log.Llongfile|log.LstdFlags|log.LUTC|log.Lmicroseconds)

	postgres := os.Getenv("PG_DB")
	if postgres == "" {
		deps.log.Println("Error setting up Postgres: no connection string set.")
		os.Exit(1)
	}
	db, err := sql.Open("postgres", postgres)
	if err != nil {
		deps.log.Printf("Error connecting to Postgres: %+v\n", err)
		os.Exit(1)
	}
	deps.db = db

	provider, err := oidc.NewProvider(context.Background(), "https://accounts.google.com")
	if err != nil {
		log.Fatal(err)
	}
	deps.verifier = provider.Verifier(&oidc.Config{
		SkipClientIDCheck: true,
		SkipNonceCheck:    true,
	})
	deps.log.Printf("Verifier: %+v\n", deps.verifier)

	http.Handle("/api/", deps.Handler("/api/"))
	err = http.ListenAndServe("0.0.0.0:4004", nil)
	if err != nil {
		deps.log.Printf("Error listening on port 0.0.0.0:4004: %+v\n", err)
		os.Exit(1)
	}
}
