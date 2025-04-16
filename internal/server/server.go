package server

import (
	"net/http"
	"regexp"
	"log"

	"github.com/gorilla/mux"
	"github.com/hmaier-dev/checklist-tool/internal/handlers"
	"github.com/hmaier-dev/checklist-tool/internal/structs"
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
	
	IndexRoute(router)

	return &Server{Router: router}
}

// Populates the NavList
func IndexRoute(router *mux.Router){
	router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		path, _ := route.GetPathTemplate()
    method, _ := route.GetMethods()
		
		// Exclude routes which do not display anything
		r, err := regexp.Compile("id")
		if err != nil {
			log.Println("Error compiling regex:", err)
			return err
		}
		match := r.MatchString(path)

		if len(method) > 0 && method[0] == "GET" && match != true{
			entry := structs.NavItem{
				Name: path,
				Path: path,
			}
			handlers.NavList = append(handlers.NavList, entry)

		}
    return nil
	})

}
