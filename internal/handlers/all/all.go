package all

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/hmaier-dev/checklist-tool/internal/database"
	"github.com/hmaier-dev/checklist-tool/internal/handlers"
)

type AllHandler struct{}

var _ handlers.DisplayHandler = (*AllHandler)(nil)

// Sets /delete and all subroutes
func (h *AllHandler)	Routes(router *mux.Router){
	router.HandleFunc("/all", h.Display).Methods("GET")
}

// Return rendered html for GET to /delete
func (h *AllHandler)	Display(w http.ResponseWriter, r *http.Request){
	var templates = []string{
		"all/templates/all.html",
		"all/templates/entries.html",
		"nav.html",
	}
  tmpl := handlers.LoadTemplates(templates)
	db := database.Init()
	all := database.GetAllEntries(db)
	view := handlers.BuildEntriesViewForTemplate()
	err := tmpl.Execute(w, map[string]any{
    "Nav" : handlers.UpdateNav(r),
		"Entries": view,
  })

  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    log.Fatalf("%q \n", err)
  }

}

func init(){
	handlers.RegisterHandler(&AllHandler{})
}
