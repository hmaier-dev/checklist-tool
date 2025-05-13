package home

import(
	"net/http"
	"log"
	"encoding/json"
	"time"

	"github.com/gorilla/mux"
	"github.com/sqids/sqids-go"

	"github.com/hmaier-dev/checklist-tool/internal/database"
	"github.com/hmaier-dev/checklist-tool/internal/handlers"
)


type HomeHandler struct{}

var _ handlers.ActionHandler = (*HomeHandler)(nil)

// Sets / and all its subroutes
func (h *HomeHandler)	Routes(router *mux.Router){
	router.HandleFunc("/", h.Display).Methods("GET")
	router.HandleFunc("/entries", h.Entries).Methods("GET")
	router.HandleFunc("/options", h.Options).Methods("GET")
	router.HandleFunc("/new", h.Execute).Methods("POST")
}

// Return html to http.ResponseWriter for /
func (h *HomeHandler) Display(w http.ResponseWriter, r *http.Request){
	var templates = []string{
		"home/templates/home.html",
		"nav.html",
		"home/templates/entries.html",
		"home/templates/options.html",
	}
  tmpl := handlers.LoadTemplates(templates)
	// This needs to be called here, to set ?template=
	// to the first template if none is set.
	db := database.Init()
	all := database.GetAllTemplates(db)
	active := ""
	if len(all) > 0 {
		active = all[0].Name
	}
	entries_raw := database.GetAllEntriesForChecklist(db, active)
	custom_fields := database.GetAllCustomFieldsForTemplate(db, active)
	entries_view := handlers.BuildEntriesViewForTemplate(custom_fields,entries_raw)
	inputs := database.GetAllCustomFieldsForTemplate(db, active)
	err := tmpl.Execute(w, map[string]any{
		"Active": active,
    "Nav" : handlers.NavList,
		"Templates": all,
		"Inputs": inputs,
		"Entries": entries_view,
  })

  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    log.Fatal("", err)
  }
}
// Loads entries per template for /
func (h *HomeHandler) Entries(w http.ResponseWriter, r *http.Request){
	template_name := r.URL.Query().Get("template")
	db := database.Init()
	entries := database.GetAllEntriesForChecklist(db, template_name)
	tmpl := handlers.LoadTemplates([]string{"home/templates/entries.html"})
	// building a map to access the descriptions by column names
	custom_fields := database.GetAllCustomFieldsForTemplate(db, template_name)
	result := handlers.BuildEntriesViewForTemplate(custom_fields, entries)
	err := tmpl.Execute(w, map[string]any{
		"Entries": result,
	})
	if err != nil{

	}
}

// Runs when submit-button on / is pressed
func (h *HomeHandler) Execute(w http.ResponseWriter, r *http.Request){
	template_name := r.FormValue("template")
	db := database.Init()
	template := database.GetChecklistTemplateByName(db, template_name)
	cols := database.GetAllCustomFieldsForTemplate(db,template_name)
	data := make(map[string]string)
	for _, col := range cols{
		key := col.Key
		value := r.FormValue(key)
		data[key] = value
	}
	json, err := json.Marshal(data)
	if err != nil{
		log.Fatalf("Error while marshaling json.\n Error: %q \n", err)
	}
	path := generatePath(data)
	entry := database.ChecklistEntry{
		Template_id: template.Id,
		Data: string(json),
		Path: path,
		Yaml: template.Empty_yaml,
		Date: time.Now().Unix(),
	}
	// Instead of checking the 'path' manually,
	// use the CONSTRAINT on the column to generate an error
	result := database.NewEntry(db, entry)
	if result != nil{
		html := `<div class='text-lime-400'>Eintrag erfolgreich erstellt.</div>`
		w.Write([]byte(html))
		return
	}else{
		html := `<div class='text-red-400'>Eintrag nicht erfolgreich erstellt.</div>`
		w.Write([]byte(html))
		return
	}
}

// Return the custom inputs fields per template
func (h *HomeHandler) Options(w http.ResponseWriter, r *http.Request){
	template_name := r.URL.Query().Get("template")
	db := database.Init()
	custom_fields := database.GetAllCustomFieldsForTemplate(db, template_name)
	tmpl := handlers.LoadTemplates([]string{"home/templates/options.html"})
	err := tmpl.Execute(w, map[string]any{
		"Inputs": custom_fields,
	})
	if err != nil{
		log.Fatalf("Can't execute 'entries'-template.\n %#v \n", err)
	}
}

func generatePath(data map[string]string) string{
	var chars []uint64
	for col := range data{
		for c := range []byte(data[col]){
			chars = append(chars,uint64(c))
		}
	}
	s, err := sqids.New()
	if err != nil{
		log.Fatalf("Error will constructing a new sqids-instance.\n Error: %q\n", err)
	}
	id, err := s.Encode(chars)
	if err != nil{
		log.Fatalf("Something wen't wrong while building the path.\n Error: %q\n", err)
	}
	return id
}

func init(){
	handlers.RegisterHandler(&HomeHandler{})
}
