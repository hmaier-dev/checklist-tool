package all

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

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
	all_templates := database.GetAllTemplates(db)
	var view []handlers.EntryView	
	for _, t := range all_templates{
		all := database.GetAllEntriesForChecklist(db,t.Name)
		for _, a := range all{
			tmp := ViewForTemplate(db, t.Name,a)
			view = append(view, tmp)
		}
	}
	err := tmpl.Execute(w, map[string]any{
    "Nav" : handlers.UpdateNav(r),
		"Entries": view,
  })

  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    log.Fatalf("%q \n", err)
  }
}

// TODO: Get the 'Template Name' from the template_id found in the database entry. Right now the information is redundant...
// Maybe connect the template_name and template_id over a JOIN()?
func ViewForTemplate(db *sql.DB, template_name string, entry database.ChecklistEntry) handlers.EntryView{
		custom_fields := database.GetAllCustomFieldsForTemplate(db,template_name)
		var fieldsMap = make(map[string]string, len(custom_fields))
		for _, field := range custom_fields{
			fieldsMap[field.Key] = field.Desc
		}
		var dataMap map[string]string
		err := json.Unmarshal([]byte(entry.Data), &dataMap)
		if err != nil{
			log.Fatalf("Error while unmarshaling json.\n Error: %q \n", err)
		}
		var viewMap []handlers.DescValueView
		// Add Template Name at first
		viewMap = append(viewMap, handlers.DescValueView{
			Desc: "Checklist",
			Value: template_name,
		})
		// Append data stored in database-entry
		for k, v := range dataMap {
				desc := fieldsMap[k]
				viewMap = append(viewMap, handlers.DescValueView{Desc: desc, Value: v, Key: k})	
		}
		// Add Erstellungsdatum as last field
		viewMap = append(viewMap, handlers.DescValueView{
			Desc: "Erstellungsdatum",
			Value: time.Unix(entry.Date,0).Format("02-01-2006 15:04:05"),
		})
		return handlers.EntryView{
			Path: entry.Path,
			Data: viewMap,			
		}
}


func init(){
	handlers.RegisterHandler(&AllHandler{})
}
