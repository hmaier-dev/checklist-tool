package main

import (
  "log"
  "net/http"
  "github.com/hmaier-dev/checklist-tool/internal/server"
)

func main() {
  server.Hello()

  srv := server.NewServer()
	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", srv.Router))
}
