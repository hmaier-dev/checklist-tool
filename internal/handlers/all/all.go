package all

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/hmaier-dev/checklist-tool/internal/database"
	"github.com/hmaier-dev/checklist-tool/internal/handlers"
	"github.com/hmaier-dev/checklist-tool/internal/server"
)

type AllHandler struct{
	Router *mux.Router	
	DB *sql.DB
}

var _ handlers.DisplayHandler = (*AllHandler)(nil)

func (h *AllHandler) New(srv *server.Server){
	h.Router = srv.Router	
	h.DB = srv.DB
}

// Sets /delete and all subroutes
func (h *AllHandler) Routes(){
	h.Router.HandleFunc("/all", h.Display).Methods("GET")
}

// Return rendered html for GET to /delete
func (h *AllHandler) Display(w http.ResponseWriter, r *http.Request){
	ctx := r.Context()
	var templates = []string{
		"all/templates/all.html",
		"all/templates/entries.html",
		"nav.html",
		"header.html",
	}
  tmpl := handlers.LoadTemplates(templates)
	var view []handlers.EntryView	
	query := database.New(h.DB)
	all, err := query.GetAllEntriesPlusTemplateName(ctx)
	for _, a := range all{
		tmp := h.ViewForTemplate(ctx, a)
		view = append(view, tmp)
	}
	err = tmpl.Execute(w, map[string]any{
		"Entries": view,
  })

  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    log.Fatalf("%q \n", err)
  }
}

// TODO: Get the 'Template Name' from the template_id found in the database entry. Right now the information is redundant...
// Maybe connect the template_name and template_id over a JOIN()?
func (h *AllHandler) ViewForTemplate(ctx context.Context, entry database.GetAllEntriesPlusTemplateNameRow) handlers.EntryView{
		var dataMap map[string]string
		err := json.Unmarshal([]byte(entry.Data), &dataMap)
		if err != nil{
			log.Fatalf("Error while unmarshaling json.\n Error: %q \n", err)
		}
		var length int = len(dataMap)
		var viewMap []handlers.DescValueView = make([]handlers.DescValueView, length)
		query := database.New(h.DB)
		custom_fields, err := query.GetCustomFieldsByTemplateName(ctx, entry.TemplateName)
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
		var t time.Time
		if entry.Date.Valid{
			t = time.Unix(entry.Date.Int64,0)		
		}else{
			t = time.Time{}
		}
		return handlers.EntryView{
			Date: t.Format("02.01.2006 15:04:05"),
			Path: entry.Path,
			Data: viewMap,
		}
}


func init(){
	handlers.RegisterHandler(&AllHandler{})
}
