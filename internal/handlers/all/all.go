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
	var view []handlers.EntryView	
	all := database.GetAllEntriesPlusTemplateName(db)
	for _, a := range all{
		tmp := ViewForTemplate(db, a)
		view = append(view, tmp)
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

// TODO: Get the 'Template Name' from the template_id found in the database entry. Right now the information is redundant...
// Maybe connect the template_name and template_id over a JOIN()?
func ViewForTemplate(db *sql.DB, entry database.EntryPlusChecklistName) handlers.EntryView{
		var dataMap map[string]string
		err := json.Unmarshal([]byte(entry.Data), &dataMap)
		if err != nil{
			log.Fatalf("Error while unmarshaling json.\n Error: %q \n", err)
		}
		var length int = len(dataMap)
		var viewMap []handlers.DescValueView = make([]handlers.DescValueView, length)
		custom_fields := database.GetAllCustomFieldsForTemplate(db,entry.TemplateName)
		var count int = 0
		// Use the order from the database-table 'custom_fields'
		for _, field := range custom_fields{
			if val, ok := dataMap[field.Key];ok{
				viewMap[count] = handlers.DescValueView{
					Desc: field.Desc, 
					Value: val, 
					Key: field.Key,
				}
			}else{
				log.Fatalf("Key '%s' not found in dataMap: '%q' \n",field.Key,dataMap)
			}
			count += 1
		}
		return handlers.EntryView{
			TemplateName: entry.TemplateName,
			Date: time.Unix(entry.Date,0).Format("02.01.2006 15:04:05"),
			Path: entry.Path,
			Data: viewMap,
		}
}


func init(){
	handlers.RegisterHandler(&AllHandler{})
}
