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
	router.HandleFunc("/checklist", handlers.Home).Methods("GET")
  sub := router.PathPrefix("/checklist").Subrouter()
  // GET
  sub.HandleFunc("/delete", handlers.DisplayDelete).Methods("Get")
  sub.HandleFunc("/reset", handlers.DisplayReset).Methods("Get")
  sub.HandleFunc("/upload", handlers.DisplayUpload).Methods("Get")
  sub.HandleFunc(`/{id:\d{14}_\d{15}}`, handlers.Display).Methods("GET")
  sub.HandleFunc(`/download_{id:\d{14}_\d{15}}`, handlers.GeneratePDF).Methods("GET")
  sub.HandleFunc("/blanko", handlers.DisplayBlanko).Methods("Get")
  
  // POST
  sub.HandleFunc("/new", handlers.NewEntry).Methods("POST")
  sub.HandleFunc(`/update_{id:\d{14}_\d{15}}`, handlers.Update).Methods("POST")
  sub.HandleFunc(`/goto_{id:\d{14}_\d{15}}`, handlers.RedirectToDownload).Methods("POST")
  sub.HandleFunc(`/delete_{id:\d{14}_\d{15}}`, handlers.DeleteEntry).Methods("POST")
  sub.HandleFunc(`/reset_{id:\d{14}_\d{15}}`, handlers.ResetChecklistForEntry).Methods("POST")
  sub.HandleFunc("/upload", handlers.ReceiveUpload).Methods("POST")

  router.PathPrefix("/checklist/static/").Handler(http.StripPrefix("/checklist/static/", http.FileServer(http.Dir("./static/"))))

	return &Server{Router: router}
}
