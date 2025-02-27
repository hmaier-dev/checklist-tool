package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/hmaier-dev/checklist-tool/internal/database"
	"github.com/hmaier-dev/checklist-tool/internal/handlers"
	"github.com/hmaier-dev/checklist-tool/internal/server"
)

func main() {
  const port = "8080"
  
  dbArg := flag.String("db", "", "Path to sqlite database")
  jsonArg := flag.String("json", "", "Path to json file representing checklist")
  flag.Parse()
  if *dbArg == "" {
    log.Fatalln("database file is mandatory")
  }
  if *jsonArg == "" {
    log.Fatalln("json checklist is mandatory")
  }

  database.DBfilePath = *dbArg
  handlers.ChecklistJsonFile = *jsonArg
  database.ChecklistJsonFile = *jsonArg

  host := fmt.Sprintf("0.0.0.0:%s", port)

  srv := server.NewServer()
	log.Printf("Starting server on %s \n", host)

  err := http.ListenAndServe(host, srv.Router)
	if err != nil {
		log.Fatal("cannot listen and server", err)
	}
}
