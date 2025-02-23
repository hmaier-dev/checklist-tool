package server

import (
  "github.com/gorilla/mux"
  "github.com/hmaier-dev/checklist-tool/internal/handlers"
)


type Server struct {
	Router *mux.Router
}

func NewServer() *Server {
	router := mux.NewRouter()
	router.HandleFunc("/health", handlers.HealthCheck).Methods("GET")
	router.HandleFunc("/checklist", handlers.CheckList).Methods("GET")
  sub := router.PathPrefix("/checklist").Subrouter()
  sub.HandleFunc("/new", handlers.New).Methods("GET")
  sub.HandleFunc(`/{id:[0-9]{15}}`, handlers.Display).Methods("GET")
  sub.HandleFunc(`/update-{id:\d{15}}`, handlers.Update).Methods("POST")

	return &Server{Router: router}
}
