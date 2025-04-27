package server

import (
	"net/http"
	"github.com/gorilla/mux"
	"github.com/hmaier-dev/checklist-tool/internal/handlers"
)


type Server struct {
	Router *mux.Router
}

func NewServer() *Server {
	router := mux.NewRouter()
	router.HandleFunc("/", handlers.Home).Methods("GET")
  sub := router.PathPrefix("/").Subrouter()
  // GET
  sub.HandleFunc("/delete", handlers.DisplayDelete).Methods("Get")
  subsub := sub.PathPrefix("/delete").Subrouter()
	subsub.HandleFunc("/entries", handlers.DeleteableEntries).Methods("GET")
  sub.HandleFunc("/reset", handlers.DisplayReset).Methods("Get")
  sub.HandleFunc("/upload", handlers.DisplayUpload).Methods("Get")
  sub.HandleFunc("/option", handlers.Options).Methods("GET")
  sub.HandleFunc("/entries", handlers.Entries).Methods("GET")
  sub.HandleFunc("/nav", handlers.Nav).Methods("GET")
	sub.HandleFunc(`/checklist/{id:\w*}`, handlers.DisplayChecklist).Methods("GET")
	sub.HandleFunc(`/print/{id:\w*}`, handlers.GeneratePDF).Methods("GET")
  
  // POST
  sub.HandleFunc("/upload", handlers.ReceiveUpload).Methods("POST")
  sub.HandleFunc("/new", handlers.NewEntry).Methods("POST")
  sub.HandleFunc("/delete", handlers.DeleteEntry).Methods("POST")

  router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	return &Server{Router: router}
}
