package checklist

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/hmaier-dev/checklist-tool/internal/database"
	"github.com/hmaier-dev/checklist-tool/internal/handlers"
	"github.com/hmaier-dev/checklist-tool/internal/pdf"
	"gopkg.in/yaml.v3"
)

// Single checkpoint of the list
type Item struct {
	Task     string           `yaml:"task"`
	Checked  bool             `yaml:"checked"`
	Children []*Item `yaml:"children,omitempty"`
	Path     string           `yaml:"Path"`
}

type ChecklistHandler struct{}

var _ handlers.DisplayHandler = (*ChecklistHandler)(nil)

func (h *ChecklistHandler)	Routes(router *mux.Router){
	router.HandleFunc(`/checklist/{id:\w*}`, h.Display).Methods("GET")
	sub := router.PathPrefix("/checklist").Subrouter()
	sub.HandleFunc(`/update/{id:\w*}`, h.Update).Methods("POST")
	sub.HandleFunc(`/print/{id:\w*}`, h.Print).Methods("GET")
}

func (h *ChecklistHandler) Display(w http.ResponseWriter, r *http.Request){
  path := mux.Vars(r)["id"]
	paths := []string{
		"checklist/templates/checklist.html",
	}
	tmpl := handlers.LoadTemplates(paths)

	db := database.Init()
	entry := database.GetEntryByPath(db, path)

	var data map[string]string
	err := json.Unmarshal([]byte(entry.Data),&data)
	if err != nil{
		log.Fatalln("Unmarshaling json from db wen't wrong.")
		return
	}
	
	template := database.GetTemplateNameByID(db,entry.Template_id)
	custom_fields := database.GetAllCustomFieldsForTemplate(db, template.Name)
	result := handlers.BuildEntryViewForTemplate(custom_fields, entry)


	// Build string for browser-tab title
	tab_desc_schema := database.GetTabDescriptionsByID(db, template.Id)
	var tab_desc string
	for i, desc := range tab_desc_schema{
		key := desc.Value
		if i == len(tab_desc_schema) - 1 {
			tab_desc += data[key]
		}else{
			tab_desc += data[key] + " | "
		}
	}
	var items []*Item
	yaml.Unmarshal([]byte(entry.Yaml), &items)
	err = tmpl.Execute(w, map[string]any{
		"TemplateName": template.Name,
		"TabDescription": tab_desc,
		"EntryView": result,
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
  alteredItem := Item{
    Task: r.Form.Get("task"),
    Checked: checked,
  }
  
  // Fetch Row from Database
  db := database.Init()
  entry := database.GetEntryByPath(db, path)

  var oldItems []*Item
  err = yaml.Unmarshal([]byte(entry.Yaml), &oldItems)
  changeCheckedStatus(alteredItem, oldItems)
  yamlBytes, err := yaml.Marshal(oldItems)

  if err != nil {
      log.Println("Error marshaling Yaml: ", err)
      return
  }
  database.UpdateYamlByPath(db, path, string(yamlBytes))
	fmt.Fprint(w,[]byte{})
}

func (h *ChecklistHandler) Print(w http.ResponseWriter, r *http.Request){
		path :=  mux.Vars(r)["id"]
		tmpl := handlers.LoadTemplates([]string{"checklist/templates/print.html"})

		db := database.Init()
		entry := database.GetEntryByPath(db, path)
		var items []Item
		err := yaml.Unmarshal([]byte(entry.Yaml), &items)

		template := database.GetTemplateNameByID(db,entry.Template_id)
		custom_fields := database.GetAllCustomFieldsForTemplate(db, template.Name)
		result := handlers.BuildEntryViewForTemplate(custom_fields, entry)

		var buf bytes.Buffer
		err = tmpl.Execute(&buf, map[string]any{
			"Items": items,
			"EntryView": result,
			"Date": time.Now().Format("02.01.2006, 15:04:05"),
		})
		bodyBytes, err := io.ReadAll(&buf)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close() // Close the body after reading
		
		var data map[string]string
		err = json.Unmarshal([]byte(entry.Data), &data)
		if err != nil{
			log.Fatalln("Unmarshaling json from db wen't wrong.")
			return
		}
		if data == nil {
			log.Fatalln("Error: data map is nil after unmarshaling.")
    return
		}
		// Build filename of pdf
		name_schema := database.GetPdfNamingByID(db, template.Id)
		
		// Build pdf name schema from entry.Data
		// Add date to the data-map because it is an extra field in the db and not present in entry.Data
		data["date"] = time.Now().Format("20060102")
		var pdfName string
		for i, desc := range name_schema{
			key := desc.Value
			if i == len(name_schema) - 1 {
				pdfName += data[key] + ".pdf"
			}else{
				pdfName += data[key] + "_"
			}
		}
		var pdfBytes []byte
		pdfBytes, err = pdf.Generate(pdfName,bodyBytes)
		if err != nil{
			log.Fatalf("Error while generating pdf.\nError: %q \n", err)
		}
		// Setting the header before sending the file to the browser
    w.Header().Set("Content-Type", "application/pdf")
		disposition := fmt.Sprintf("attachment; filename=%s", pdfName)
    w.Header().Set("Content-Disposition", disposition)
    _, err = io.Copy(w, bytes.NewReader(pdfBytes))
    if err != nil {
			log.Fatalf("Couldn't send pdf to browser.\nError: %q \n", err)
    }
}

func changeCheckedStatus(newItem Item, oldChecklist []*Item){
  for _, item := range oldChecklist{
    if newItem.Task == item.Task{
      item.Checked = newItem.Checked     
      return
    }
    if len(item.Children) > 0 {
			changeCheckedStatus(newItem, item.Children)
		}
  }
}



func init(){
	handlers.RegisterHandler(&ChecklistHandler{})
}
