package server

import (
  "fmt"
  "github.com/gorilla/mux"
  "github.com/hmaier-dev/checklist-tool/internal/handlers"
)

func Hello(){
  fmt.Println("Hello from Server!")
}

type Server struct {
	Router *mux.Router
}

func NewServer() *Server {
	router := mux.NewRouter()
	router.HandleFunc("/health", handlers.HealthCheck).Methods("GET")
	return &Server{Router: router}
}
