package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/hmaier-dev/checklist-tool/internal/structs"
	"github.com/hmaier-dev/checklist-tool/internal/database"
	"github.com/hmaier-dev/checklist-tool/internal/handlers"
	"github.com/hmaier-dev/checklist-tool/internal/server"

	"gopkg.in/yaml.v3"
)


//go:embed checklists/*
var embedFS embed.FS


func main() {
  const port = "8080"
  
  dbArg := flag.String("db", "", "Path to sqlite database")
  flag.Parse()
  if *dbArg == "" {
    flag.Usage()
    log.Fatalln("database file is mandatory")
  }

  database.DBfilePath = *dbArg

  // Get a raw/blank checklist
  clDefaultFile, err := embedFS.ReadFile("checklists/default.yml")
  if err != nil {
    log.Fatalf("could not read embedded file: ", err)
  }

	var clDefault []*structs.ChecklistItem
	err = yaml.Unmarshal(clDefaultFile, &clDefault)
	if err != nil {
		fmt.Println("Error parsing YAML:", err)
		return
	}

  // Pass empty checklist as raw file and struct
  database.EmptyChecklist = clDefaultFile
  database.EmptyChecklistItemsArray = clDefault
  handlers.EmptyChecklist = clDefaultFile

  host := fmt.Sprintf("0.0.0.0:%s", port)

  srv := server.NewServer()
	log.Printf("Starting server on %s \n", host)

  err = http.ListenAndServe(host, srv.Router)
	if err != nil {
		log.Fatal("cannot listen and server", err)
	}
}
