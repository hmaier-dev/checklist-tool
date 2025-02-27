package main

import (
	"fmt"
	"log"
	"net/http"
  "flag"

	"github.com/hmaier-dev/checklist-tool/internal/server"
	"github.com/hmaier-dev/checklist-tool/internal/database"
)

func main() {
  const port = "8080"
  
  dbArg := flag.String("db", "", "Path to sqlite database")
  flag.Parse()
  if *dbArg == "" {
    log.Fatalln("Database File is mandatory")
  }

  database.DBfilePath = *dbArg

  host := fmt.Sprintf("0.0.0.0:%s", port)

  srv := server.NewServer()
	log.Printf("Starting server on %s \n", host)

  err := http.ListenAndServe(host, srv.Router)
	if err != nil {
		log.Fatal("cannot listen and server", err)
	}
}
