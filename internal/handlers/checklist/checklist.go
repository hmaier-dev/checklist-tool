package checklist

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/hmaier-dev/checklist-tool/internal/database"
	"github.com/hmaier-dev/checklist-tool/internal/handlers"
	"github.com/hmaier-dev/checklist-tool/internal/pdf"
	"github.com/hmaier-dev/checklist-tool/internal/server"
	"gopkg.in/yaml.v3"
)

// Single checkpoint of the list
type Item struct {
	Task     string  `yaml:"task"`
	Checked  bool    `yaml:"checked"`
	Text     *string `yaml:"text"` // this needs to be a pointer, because that way {{ if .Text }} displays input fields, even with an empty string
	Children []*Item `yaml:"children,omitempty"`
	Path     string  `yaml:"Path"`
}

type ChecklistHandler struct{
	Router *mux.Router	
	DB *sql.DB
}

var _ handlers.DisplayHandler = (*ChecklistHandler)(nil)

func (h *ChecklistHandler) New(srv *server.Server){
	h.Router = srv.Router	
	h.DB = srv.DB
}

func (h *ChecklistHandler) Routes(){
	h.Router.HandleFunc(`/checklist/{id:\w*}`, h.Display).Methods("GET")
	sub := h.Router.PathPrefix("/checklist").Subrouter()
	sub.HandleFunc(`/update/check/{id:\w*}`, h.UpdateCheckedState).Methods("POST")
	sub.HandleFunc(`/update/text/{id:\w*}`, h.UpdateText).Methods("POST")
	sub.HandleFunc(`/print/{id:\w*}`, h.Print).Methods("GET")
	sub.HandleFunc("/delete", h.Delete).Methods("POST")
}

func (h *ChecklistHandler) Display(w http.ResponseWriter, r *http.Request){
	ctx := r.Context()
  path := mux.Vars(r)["id"]
	paths := []string{
		"checklist/templates/checklist.html",
		"nav.html",
		"header.html",
		"history/templates/history.html",
	}
	tmpl := handlers.LoadTemplates(paths)

	q := database.New(h.DB)
	entry, err := q.GetEntryByPath(ctx, path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var data map[string]string
	err = json.Unmarshal([]byte(entry.Data),&data)
	if err != nil{
		log.Fatalln("Unmarshaling json from db wen't wrong.")
		return
	}
	
	templateName, err := q.GetTemplateNameById(ctx, entry.TemplateID)
	customFields, err := q.GetCustomFieldsByTemplateName(ctx, templateName)
	result := handlers.BuildEntryViewForTemplate(customFields, &entry)

	// Build string for browser-tab title
	tab_desc_schema, err := q.GetTabDescriptionsByTemplateID(ctx, entry.TemplateID)
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
	var y string
	if entry.Yaml.Valid{
		y = entry.Yaml.String
	}else{
		log.Printf("%#v \n", entry.Yaml)
		http.Error(w, "Checklist will not render correctly. 'yaml'-field wasn't valid.", http.StatusInternalServerError)
		return
	}
	yaml.Unmarshal([]byte(y), &items)
	err = tmpl.Execute(w, map[string]any{
		"TemplateName": templateName,
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

func (h *ChecklistHandler) UpdateCheckedState(w http.ResponseWriter, r *http.Request){
	ctx := r.Context()
	path :=  mux.Vars(r)["id"]
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}
  var checked bool = false
  // if ["checked"] isset
  if _, ok := r.Form["checked"]; ok{
    checked = true
  }
	var alteredItem Item
	alteredItem = Item{
		Task: r.Form.Get("task"),
		Text: nil,
		Checked: checked,
	}
  // Fetch Row from Database
	q := database.New(h.DB)
  entry, err := q.GetEntryByPath(ctx, path)
	if err != nil {
		http.Error(w,err.Error(), http.StatusInternalServerError)
		return
	}
  var oldItems []*Item
	var y string
	if entry.Yaml.Valid{
		y = entry.Yaml.String
	}else{
		msg := "Checkpoint state couldn't get updated. 'yaml'-field wasn't valid."
		log.Println(msg)
		http.Error(w, msg, http.StatusInternalServerError)
	}
  err = yaml.Unmarshal([]byte(y), &oldItems)
  alterCheckedState(alteredItem, oldItems)
  yamlBytes, err := yaml.Marshal(oldItems)
  if err != nil {
      log.Println("Error marshaling Yaml: ", err)
      return
  }
	
	arg := database.UpdateYamlByPathParams{
		Yaml: sql.NullString{Valid: true, String: string(yamlBytes)},
		Path: path,
	}

  q.UpdateYamlByPath(ctx, arg)
	w.Write([]byte{})
}

func alterCheckedState(newItem Item, checklistSlice []*Item){
  for _, item := range checklistSlice{
		// The first occurence of a task is altered.
		// This way 'Item.Task' should be unique. Otherwise it cannot get altered.
    if newItem.Task == item.Task{
				item.Checked = newItem.Checked
      return
    }
		// Go into the lower levels
    if len(item.Children) > 0 {
			alterCheckedState(newItem, item.Children)
		}
  }
}

func (h *ChecklistHandler) UpdateText(w http.ResponseWriter, r *http.Request){
	ctx := r.Context()
	path :=  mux.Vars(r)["id"]
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}
	text :=  r.Form.Get("text")
	var alteredItem Item
	alteredItem = Item{
		Task: r.Form.Get("task"),
		Text: &text,
	}
	q := database.New(h.DB)
  entry, err := q.GetEntryByPath(ctx, path)
	if err != nil {
		http.Error(w,err.Error(), http.StatusInternalServerError)
		return
	}
  var oldItems []*Item
	var y string
	if entry.Yaml.Valid{
		y = entry.Yaml.String
	}else{
		msg := "Text field couldn't get updated. 'yaml'-field wasn't valid."
		log.Println(msg)
		http.Error(w, msg, http.StatusInternalServerError)
	}
  err = yaml.Unmarshal([]byte(y), &oldItems)
  updateTextState(alteredItem, oldItems)
  yamlBytes, err := yaml.Marshal(oldItems)

  if err != nil {
      log.Println("Error marshaling Yaml: ", err)
      return
  }

	arg := database.UpdateYamlByPathParams{
		Yaml: sql.NullString{Valid: true, String: string(yamlBytes)},
		Path: path,
	}

  q.UpdateYamlByPath(ctx, arg)
	w.Write([]byte{})
}

func updateTextState(newItem Item, checklistSlice []*Item){
  for _, item := range checklistSlice{
    if newItem.Task == item.Task{
			*item.Text = *newItem.Text
      return
    }
		// Go into the lower levels
    if len(item.Children) > 0 {
			updateTextState(newItem, item.Children)
		}
  }
}

func (h *ChecklistHandler) Print(w http.ResponseWriter, r *http.Request){
	ctx := r.Context()
	path :=  mux.Vars(r)["id"]
	tmpl := handlers.LoadTemplates([]string{"checklist/templates/print.html"})

	q := database.New(h.DB)
	entry, err := q.GetEntryByPath(ctx, path)
	var items []Item
	var y string
	if entry.Yaml.Valid{
		y = entry.Yaml.String
	}else{
		msg := "Couldn't render the document. 'yaml'-field wasn't valid."
		log.Println(msg)
		http.Error(w, msg, http.StatusInternalServerError)
	}
	err = yaml.Unmarshal([]byte(y), &items)

	templateName, err := q.GetTemplateNameById(ctx, entry.TemplateID)
	customFields, err := q.GetCustomFieldsByTemplateName(ctx, templateName)
	result := handlers.BuildEntryViewForTemplate(customFields, &entry)
	
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
	
	// Build pdf_name_schema from entry.Data
	// Add date to the data-map because it is an extra field in the db and not present in entry.Data
	nameSchema, err := q.GetPdfNamingByTemplateID(ctx, entry.TemplateID)
	data["date"] = time.Now().Format("20060102")
	var pdfName string
	for i, desc := range nameSchema{
		key := desc.Value
		if i == len(nameSchema) - 1 {
			pdfName += data[key] + ".pdf"
		}else{
			// Removes all <spaces> in the file name
			val := strings.ReplaceAll(data[key], " ", "_")
			pdfName += val + "_"
		}
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, map[string]any{
		"Title": pdfName,
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
	// Reponse from gotenberg api
	response, err := pdf.Generate(r, pdfName, bodyBytes)
	if err != nil {
		log.Printf("Couldn't send pdf to browser.\nError: %q \n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return 
	}
	defer response.Body.Close()

	// Setting the header before sending the file to the browser
	w.Header().Set("Content-Type", "application/pdf")
	disposition := fmt.Sprintf("attachment; filename=%s", pdfName)
	w.Header().Set("Content-Disposition", disposition)

	_, err = io.Copy(w, response.Body)
	if err != nil {
		log.Fatalf("Couldn't send pdf to browser.\nError: %q \n", err)
	}
}

func (h *ChecklistHandler) Delete(w http.ResponseWriter, r *http.Request){
	ctx := r.Context()
	path := r.FormValue("path")
	q := database.New(h.DB)
	q.DeleteEntryByPath(ctx,path)

	// Special header for htmx
	w.Header().Set("HX-Redirect", "/all")
	w.WriteHeader(http.StatusNoContent)
}


func init(){
	handlers.RegisterHandler(&ChecklistHandler{})
}
