package reset

import (
	"fmt"
	"log"
	"net/http"
	"github.com/gorilla/mux"
	"github.com/hmaier-dev/checklist-tool/internal/database"
	"github.com/hmaier-dev/checklist-tool/internal/handlers"
)


type ResetHandler struct {}

// ResetHandler must implement all methods used in handlers.Handler.
// Because go doesn't have a implements, I test it with this declare.
// In go interface implementation is implicit, meaning you
// don't need to write it out
var _ handlers.ActionHandler = (*ResetHandler)(nil)


func (h *ResetHandler) Routes(router *mux.Router){
	sub := router.PathPrefix("/reset").Subrouter()
  sub.HandleFunc("", h.Display).Methods("Get")
	sub.HandleFunc("/entries", h.Entries).Methods("GET")
  sub.HandleFunc("", h.Execute).Methods("POST")
}


func (h *ResetHandler) Display(w http.ResponseWriter, r *http.Request){
	var templates = []string{
		"reset/templates/reset.html",
		"nav.html",
	}
  tmpl := handlers.LoadTemplates(templates)
	db := database.Init()
	all := database.GetAllTemplates(db)
	active := r.URL.Query().Get("template")
	err := tmpl.Execute(w, map[string]any{
		"Active": active,
		"Nav": handlers.UpdateNav(r),
		"Templates": all,
  })
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    log.Fatal("", err)
  }

}
func (h *ResetHandler) Entries(w http.ResponseWriter, r *http.Request){
	var templates = []string{
		"reset/templates/entries.html",
	}
  tmpl := handlers.LoadTemplates(templates)
	db := database.Init()
	active := r.URL.Query().Get("template")
	entries_raw := database.GetAllEntriesForChecklist(db, active)
	custom_fields := database.GetAllCustomFieldsForTemplate(db, active)
	entries_view := handlers.BuildEntriesView(custom_fields,entries_raw)
	err := tmpl.Execute(w, map[string]any{
		"Entries": entries_view,
  })
	if err != nil{
		fmt.Fprintf(w,"Can't execute 'entries'-template.\n %#v \n", err)
	}
	if err != nil{
		log.Fatalf("Something went wrong executing the 'home/templates/entries.html'-template.\n %q \n", err)
	}
}
func (h *ResetHandler) Execute(w http.ResponseWriter, r *http.Request){
  http.Redirect(w, r, "/reset", http.StatusSeeOther)
}

func init(){
	// to trigger this, import the package blank
	// e.g. _ "github.com/tool/handler/pkg"
	handlers.RegisterHandler(&ResetHandler{})	
}
