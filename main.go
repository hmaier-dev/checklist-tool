package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/hmaier-dev/checklist-tool/internal/server"
)

func main() {
  const port = "8080"
  host := fmt.Sprintf("0.0.0.0:%s", port)

  srv := server.NewServer()
	log.Printf("Starting server on %s \n", host)

  err := http.ListenAndServe(host, srv.Router)
	if err != nil {
		log.Fatal("cannot listen and server", err)
	}
}
