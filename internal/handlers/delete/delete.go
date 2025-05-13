package delete

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/hmaier-dev/checklist-tool/internal/database"
	"github.com/hmaier-dev/checklist-tool/internal/handlers"
)

type DeleteHandler struct{}

var _ handlers.ActionHandler = (*DeleteHandler)(nil)

// Sets /delete and all subroutes
func (h *DeleteHandler)	Routes(router *mux.Router){
	sub := router.PathPrefix("/delete").Subrouter()
	sub.HandleFunc("", h.Display).Methods("GET")
	sub.HandleFunc("/entries", h.Entries).Methods("GET")
	sub.HandleFunc("", h.Execute).Methods("POST")

}

// Return rendered html for GET to /delete
func (h *DeleteHandler)	Display(w http.ResponseWriter, r *http.Request){
	var templates = []string{
		"delete/templates/delete.html",
		"delete/templates/entries.html",
		"nav.html",
	}
  tmpl := handlers.LoadTemplates(templates)
	db := database.Init()
	entries := database.GetAllEntriesPlusTemplateName(db)
	var view []handlers.EntryView = make([]handlers.EntryView, len(entries))
	for i, entry := range entries{
		view[i] = handlers.ViewForEntry(db, entry)
	}
	err := tmpl.Execute(w, map[string]any{
    "Nav" : handlers.NavList,
		"Entries": view,
  })

  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    log.Fatalf("%q \n", err)
  }

}

// Returns all entries for a 'template' into deleteEntries.html
func (h *DeleteHandler)	Entries(w http.ResponseWriter, r *http.Request){
	var templates = []string{
		"delete/templates/entries.html",
	}
  tmpl := handlers.LoadTemplates(templates)
	db := database.Init()
	entries := database.GetAllEntriesPlusTemplateName(db)
	var view []handlers.EntryView = make([]handlers.EntryView, len(entries))
	for i, entry := range entries{
		view[i] = handlers.ViewForEntry(db, entry)
	}
	err := tmpl.Execute(w, map[string]any{
		"Entries": view,
	})
	if err != nil{
		fmt.Fprintf(w,"Error while load template.\n %q \n", err)
	}
}

// Removes entry from 'entries'-table by the 'path'-column
func (h *DeleteHandler)	Execute(w http.ResponseWriter, r *http.Request){
	path := r.FormValue("path")
	db := database.Init()
	defer db.Close()
	database.DeleteEntryByPath(db,path)

	// Special header for htmx
	w.Header().Set("HX-Redirect", "/delete")
	w.WriteHeader(http.StatusNoContent)
}

func init(){
	handlers.RegisterHandler(&DeleteHandler{})
}
