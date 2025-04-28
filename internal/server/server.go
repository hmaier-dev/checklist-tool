package server

import (
	"net/http"
	"github.com/gorilla/mux"
	"github.com/hmaier-dev/checklist-tool/internal/handlers"
	// blank import for handlers. They initalize theirself by init()
	_ "github.com/hmaier-dev/checklist-tool/internal/handlers/reset"
	_ "github.com/hmaier-dev/checklist-tool/internal/handlers/delete"
)


type Server struct {
	Router *mux.Router
}

func NewServer() *Server {
	router := mux.NewRouter()
	router.HandleFunc("/", handlers.Home).Methods("GET")
  sub := router.PathPrefix("/").Subrouter()

  sub.HandleFunc("/upload", handlers.DisplayUpload).Methods("Get")
  sub.HandleFunc("/option", handlers.Options).Methods("GET")
  sub.HandleFunc("/entries", handlers.Entries).Methods("GET")

  sub.HandleFunc("/nav", handlers.Nav).Methods("GET")
	sub.HandleFunc(`/checklist/{id:\w*}`, handlers.DisplayChecklist).Methods("GET")
	sub.HandleFunc(`/print/{id:\w*}`, handlers.GeneratePDF).Methods("GET")
  
  // POST
  sub.HandleFunc("/upload", handlers.ReceiveUpload).Methods("POST")
  sub.HandleFunc("/new", handlers.NewEntry).Methods("POST")

	for _, h := range handlers.GetHandlers() {
		// Link the routes declared in sub-handlers to *mux.Router
		h.Routes(sub)
	}

  router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	return &Server{Router: router}
}
