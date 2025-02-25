package server

import (
  "github.com/gorilla/mux"
  "github.com/hmaier-dev/checklist-tool/internal/handlers"
  "net/http"
)


type Server struct {
	Router *mux.Router
}

func NewServer() *Server {
	router := mux.NewRouter()
	router.HandleFunc("/checklist", handlers.Home).Methods("GET")
  sub := router.PathPrefix("/checklist").Subrouter()
  sub.HandleFunc("/new", handlers.NewEntry).Methods("POST")
  sub.HandleFunc("/blanko", handlers.DisplayBlanko).Methods("Get")
  sub.HandleFunc(`/{id:[0-9]{15}}`, handlers.Display).Methods("GET")
  sub.HandleFunc(`/update_{id:[0-9]{15}}`, handlers.Update).Methods("POST")
  router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))


	return &Server{Router: router}
}
