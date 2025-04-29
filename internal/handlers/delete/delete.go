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
  wd, err := os.Getwd()
  if err != nil{
    log.Fatal("couldn't get working directory: ", err)
  }
	var static = filepath.Join(wd, "static")
	var delete_tmpl = filepath.Join(static, "delete.html")
	var nav_tmpl = filepath.Join(static, "nav.html")

  tmpl := template.Must(template.ParseFiles(delete_tmpl, nav_tmpl))
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    log.Fatal("error parsing home template: ", err)
  }
	db := database.Init()
	all := database.GetAllTemplates(db)
  err = tmpl.Execute(w, map[string]any{
    "Nav" : handlers.UpdateNav(r),
		"Templates": all,
		"Active": r.URL.Query().Get("template"),
  })

  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    log.Fatalf("%q \n", err)
  }

}

// Returns all entries for a 'template' into deleteEntries.html
func (h *DeleteHandler)	Entries(w http.ResponseWriter, r *http.Request){
	template_name := r.URL.Query().Get("template")
  wd, err := os.Getwd()
  if err != nil{
    log.Fatal("couldn't get working directory: ", err)
  }
	var entries_tmpl = filepath.Join(wd, "static/deleteEntries.html")
	tmpl := template.Must(template.ParseFiles(entries_tmpl))

	db := database.Init()
	custom_fields := database.GetAllCustomFieldsForTemplate(db, template_name)
	entries := database.GetAllEntriesForChecklist(db, template_name)
	result := handlers.BuildEntriesView(custom_fields, entries)
	err = tmpl.Execute(w, map[string]any{
		"Entries": result,
	})
}

// Removes entry from 'entries'-table by the 'path'-column
func (h *DeleteHandler)	Execute(w http.ResponseWriter, r *http.Request){
  http.Redirect(w, r, "/delete", http.StatusSeeOther)
}

func init(){
	handlers.RegisterHandler(&DeleteHandler{})
}
