package delete

import (
	"fmt"
	"log"
	"net/http"
	"database/sql"

	"github.com/gorilla/mux"

	"github.com/hmaier-dev/checklist-tool/internal/database"
	"github.com/hmaier-dev/checklist-tool/internal/handlers"
	"github.com/hmaier-dev/checklist-tool/internal/server"
)

type DeleteHandler struct{
	Router *mux.Router
	DB *sql.DB
}

var _ handlers.ActionHandler = (*DeleteHandler)(nil)

func (h *DeleteHandler) New(srv *server.Server){
	h.Router = srv.Router	
	h.DB = srv.DB
}

// Sets /delete and all subroutes
func (h *DeleteHandler)	Routes(){
	sub := h.Router.PathPrefix("/delete").Subrouter()
	sub.HandleFunc("", h.Display).Methods("GET")
	sub.HandleFunc("/entries", h.Entries).Methods("GET")
	sub.HandleFunc("", h.Execute).Methods("POST")

}

// Return rendered html for GET to /delete
func (h *DeleteHandler)	Display(w http.ResponseWriter, r *http.Request){
	ctx := r.Context()
	var templates = []string{
		"delete/templates/delete.html",
		"delete/templates/entries.html",
		"nav.html",
		"header.html",
	}
  tmpl := handlers.LoadTemplates(templates)
	q := database.New(h.DB)
	entries, err := q.GetAllEntriesPlusTemplateName(ctx)
	var view []handlers.EntryView = make([]handlers.EntryView, len(entries))
	for i, entry := range entries{
		view[i] = handlers.ViewForEntry(h.DB, ctx, entry)
	}
	err = tmpl.Execute(w, map[string]any{
		"Entries": view,
  })

  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    log.Fatalf("%q \n", err)
  }

}

// Returns all entries for a 'template' into deleteEntries.html
func (h *DeleteHandler)	Entries(w http.ResponseWriter, r *http.Request){
	ctx := r.Context()
	var templates = []string{
		"delete/templates/entries.html",
	}
  tmpl := handlers.LoadTemplates(templates)
	q := database.New(h.DB)
	entries, err := q.GetAllEntriesPlusTemplateName(ctx)
	if err != nil{
		msg := "Couldn't load all entries plus template name."
		log.Println(msg)
		http.Error(w,msg,http.StatusInternalServerError)
		return
	}
	var view []handlers.EntryView = make([]handlers.EntryView, len(entries))
	for i, entry := range entries{
		view[i] = handlers.ViewForEntry(h.DB, ctx, entry)
	}
	err = tmpl.Execute(w, map[string]any{
		"Entries": view,
	})
	if err != nil{
		fmt.Fprintf(w,"Error while load template.\n %q \n", err)
	}
}

// Removes entry from 'entries'-table by the 'path'-column
func (h *DeleteHandler)	Execute(w http.ResponseWriter, r *http.Request){
	ctx := r.Context()
	path := r.FormValue("path")
	q := database.New(h.DB)
	q.DeleteEntryByPath(ctx,path)

	// Special header for htmx
	w.Header().Set("HX-Redirect", "/delete")
	w.WriteHeader(http.StatusNoContent)
}

func init(){
	handlers.RegisterHandler(&DeleteHandler{})
}
