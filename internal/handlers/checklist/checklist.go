package checklist

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/hmaier-dev/checklist-tool/internal/database"
	"github.com/hmaier-dev/checklist-tool/internal/handlers"
	"gopkg.in/yaml.v3"
)

// Single checkpoint of the list
type ChecklistItem struct {
	Task     string           `yaml:"task"`
	Checked  bool             `yaml:"checked"`
	Children []*ChecklistItem `yaml:"children,omitempty"`
	Path     string           `yaml:"Path"`
}

type ChecklistHandler struct{}

var _ handlers.DisplayHandler = (*ChecklistHandler)(nil)

func (h *ChecklistHandler)	Routes(router *mux.Router){
	router.HandleFunc(`/checklist/{id:\w*}`, h.Display).Methods("GET")
	sub := router.PathPrefix("/checklist").Subrouter()
	sub.HandleFunc(`/update/{id:\w*}`, h.Update).Methods("GET")
}

func (h *ChecklistHandler) Display(w http.ResponseWriter, r *http.Request){
  path := mux.Vars(r)["id"]
	paths := []string{
		"checklist/templates/checklist.html",
	}
	tmpl := handlers.LoadTemplates(paths)

	db := database.Init()
	entries := make([]database.ChecklistEntry, 1)
	entries[0] = database.GetEntryByPath(db, path)
	template := database.GetTemplateNameByID(db,entries[0].Template_id)
	custom_fields := database.GetAllCustomFieldsForTemplate(db, template.Name)
	result := handlers.BuildEntriesView(custom_fields, entries)

	var items []*ChecklistItem
	yaml.Unmarshal([]byte(entries[0].Yaml), &items)
	err := tmpl.Execute(w, map[string]any{
		"Entries": result,
		"Items": items,
		"Path": path,
  })
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    log.Fatal("", err)
  }
}

func (h *ChecklistHandler) Update(w http.ResponseWriter, r *http.Request){
	path :=  mux.Vars(r)["id"]
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}
  
  var checked bool
  // if ["checked"] isset
  if _, ok := r.Form["checked"]; ok{
    checked = true
  }else{
    checked = false
  }
  alteredItem := ChecklistItem{
    Task: r.Form.Get("task"),
    Checked: checked,
  }
  
  // Fetch Row from Database
  db := database.Init()
  entry := database.GetEntryByPath(db, path)
  if err != nil{
    log.Fatalf("error fetching data by path: %q", err)
  }

  var oldItems []*ChecklistItem
  err = yaml.Unmarshal([]byte(entry.Yaml), &oldItems)
  ChangeCheckedStatus(alteredItem, oldItems)
  yamlBytes, err := yaml.Marshal(oldItems)

  if err != nil {
      log.Println("Error marshaling Yaml: ", err)
      return
  }
  database.UpdateYamlByPath(db, path, string(yamlBytes))
}

func ChangeCheckedStatus(newItem ChecklistItem, oldChecklist []*ChecklistItem){
  for _, item := range oldChecklist{
    if newItem.Task == item.Task{
      item.Checked = newItem.Checked     
      return
    }
    if len(item.Children) > 0 {
			ChangeCheckedStatus(newItem, item.Children)
		}
  }
}



func init(){
	handlers.RegisterHandler(&ChecklistHandler{})
}
