package server

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gorilla/mux"

)

type Server struct {
	Router *mux.Router
	DB *sql.DB
}

func NewServer(db *sql.DB) *Server {
	router := mux.NewRouter()
	srv := &Server{
		Router: router,
		DB: db,
	}
  router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	// logs all routes when starting after they go defined
	router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		path, _ := route.GetPathTemplate()
		method, _ := route.GetMethods()
		log.Println(method, path)
		return nil
	})

	return srv
}
