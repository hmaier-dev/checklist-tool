package new

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/btcsuite/btcutil/base58"
	"github.com/gorilla/mux"

	"github.com/hmaier-dev/checklist-tool/internal/database"
	"github.com/hmaier-dev/checklist-tool/internal/handlers"
	"github.com/hmaier-dev/checklist-tool/internal/server"
)


type NewHandler struct{
	Router *mux.Router	
	DB *sql.DB
}

var _ handlers.DisplayHandler = (*NewHandler)(nil)

func (h *NewHandler) New(srv *server.Server){
	h.Router = srv.Router	
	h.DB = srv.DB
}

// Sets / and all its subroutes
func (h *NewHandler)	Routes(){
	h.Router.HandleFunc("/", h.Display).Methods("GET")
	h.Router.HandleFunc("/entries", h.Entries).Methods("GET")
	h.Router.HandleFunc("/options", h.Options).Methods("GET")
	h.Router.HandleFunc("/new", h.Execute).Methods("POST")
}

// Return html to http.ResponseWriter for /
func (h *NewHandler) Display(w http.ResponseWriter, r *http.Request){
	ctx := r.Context()
	var templates = []string{
		"new/templates/new.html",
		"new/templates/entries.html",
		"new/templates/options.html",
		"nav.html",
		"header.html",
	}
  tmpl := handlers.LoadTemplates(templates)
	// This needs to be called here, to set ?template=
	// to the first template if none is set.
	q := database.New(h.DB)
	all, err := q.GetAllTemplates(ctx)
	if err != nil{
		msg := "Couldn't load all templates"
		log.Println(msg)
		http.Error(w,msg,http.StatusInternalServerError)
	}
	active := ""
	if len(all) > 0 {
		active = all[0].Name
	}
	entriesActiveTemplate, err := q.GetEntriesByTemplateName(ctx, active)
	customFields, err := q.GetCustomFieldsByTemplateName(ctx, active)
	entriesView := handlers.BuildEntriesViewForTemplate(customFields, entriesActiveTemplate)

	err = tmpl.Execute(w, map[string]any{
		"Active": active,
		"Templates": all,
		"Inputs": customFields,
		"Entries": entriesView,
  })

  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
}
// Loads entries per template for /
func (h *NewHandler) Entries(w http.ResponseWriter, r *http.Request){
	ctx := r.Context()
	templateName := r.URL.Query().Get("template")
	q := database.New(h.DB)
	entries, err := q.GetEntriesByTemplateName(ctx, templateName)
	tmpl := handlers.LoadTemplates([]string{"new/templates/entries.html"})
	// building a map to access the descriptions by column names
	customFields, err := q.GetCustomFieldsByTemplateName(ctx, templateName)
	result := handlers.BuildEntriesViewForTemplate(customFields, entries)
	err = tmpl.Execute(w, map[string]any{
		"Entries": result,
	})
	if err != nil{
		msg := fmt.Sprintf("Couldn't load entries for template: '%s'.", templateName)
		log.Println(msg)
		http.Error(w,msg,http.StatusInternalServerError)
	}
}

// Runs when submit-button on / is pressed
func (h *NewHandler) Execute(w http.ResponseWriter, r *http.Request){
	ctx := r.Context()
	templateName := r.FormValue("template")
	q := database.New(h.DB)
	template, err := q.GetTemplateByName(ctx, templateName)
	//TODO: check if no template in db trigger this
	if err != nil{
		html := `<div class='text-red-700'>Da keine Checkliste verfügbar ist, kann kein Eintrag angelegt werden.</div>`
		w.Write([]byte(html))
		return
	}
	cols, err := q.GetCustomFieldsByTemplateName(ctx,templateName)
	data := make(map[string]string)
	for _, col := range cols{
		// Only read keys from the form,
		// which have been specified in 'custom_fields' database schema.
		// That way, no invalid data can be passed
		key := col.Key
		value := r.FormValue(key)
		data[key] = value
	}
	json, err := json.Marshal(data)
	if err != nil{
		msg := fmt.Sprintf("Error while marshaling json.\n Error: %q \n", err)
		log.Println(msg)
		http.Error(w,msg,http.StatusInternalServerError)
	}
	path := generatePath(data)
	params := database.InsertEntryParams{
		TemplateID: template.ID,
		Data: string(json),
		Path: path,
		Yaml: template.EmptyYaml,
		Date: sql.NullInt64{Valid: true, Int64: time.Now().Unix()},
	}
	// Instead of checking the 'path' manually,
	// use the CONSTRAINT on the column to generate an error
	err = q.InsertEntry(ctx, params)

	if err != nil{
		switch err.Error(){
		case "UNIQUE constraint failed: entries.path":
			html := `<div class='text-red-700'>Eintrag ist bereits vorhanden und wurde daher nicht erneut erstellt.</div>`
			w.Write([]byte(html))
			return
		default:
			html := `<div class='text-red-700'>Ein unbekannter Fehler aufgetreten.</div>`
			w.Write([]byte(html))
			return
		}
	}else{
		html := `<div class='text-emerald-600'>Eintrag erfolgreich erstellt.</div>`
		w.Write([]byte(html))
		return
	}
}

// Return the custom inputs fields per template
func (h *NewHandler) Options(w http.ResponseWriter, r *http.Request){
	ctx := r.Context()
	templateName := r.URL.Query().Get("template")
	q := database.New(h.DB)
	customFields, err := q.GetCustomFieldsByTemplateName(ctx, templateName)
	if err != nil{
		msg := fmt.Sprintf("Couldn't get template '%s' for rendering options", templateName)
		log.Println(msg)
		http.Error(w,msg,http.StatusInternalServerError)
	}
	tmpl := handlers.LoadTemplates([]string{"new/templates/options.html"})
	err = tmpl.Execute(w, map[string]any{
		"Inputs": customFields,
	})
	if err != nil{
		msg := "Couldn't render options template."
		log.Println(msg)
		http.Error(w,msg,http.StatusInternalServerError)
	}
}

func generatePath(data map[string]string) string{
	keys := make([]string, 0, len(data))
	for k := range data{
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var chars []byte
	for _, k := range keys{
		chars = append(chars, []byte(data[k])...)
	}
	algo := sha256.New()
	algo.Write(chars)
	// base58 is used, because it misses certain chars which makes it more human-readable.
	// It's also used for bitcoin-address and stuff like that
	id := base58.Encode(algo.Sum(nil))
	return id
}

func init(){
	handlers.RegisterHandler(&NewHandler{})
}
