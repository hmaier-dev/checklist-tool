package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/hmaier-dev/checklist-tool/internal/database"
	"github.com/hmaier-dev/checklist-tool/internal/server"
)



func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
  const port = "8080"
  
  dbArg := flag.String("db", "", "Path to sqlite database")
  flag.Parse()
  if *dbArg == "" {
    flag.Usage()
    log.Fatalln("database file is mandatory")
  }

  database.DBfilePath = *dbArg

  host := fmt.Sprintf("0.0.0.0:%s", port)
  srv := server.NewServer()
	log.Printf("Starting tool on %s \n", host)
	err := http.ListenAndServe(host, srv.Router)
	if err != nil {
		log.Fatal("cannot listen and server", err)
	}
}
