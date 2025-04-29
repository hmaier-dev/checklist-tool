package handlers

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hmaier-dev/checklist-tool/internal/database"
	"github.com/hmaier-dev/checklist-tool/internal/pdf"
	"github.com/hmaier-dev/checklist-tool/internal/structs"

	"github.com/gorilla/mux"
	"gopkg.in/yaml.v3"
)

type Handler interface{
	Routes(router *mux.Router)
	Display(w http.ResponseWriter, r *http.Request)	
	Entries(w http.ResponseWriter, r *http.Request)	
	Execute(w http.ResponseWriter, r *http.Request)
}

var handlerRegistry []Handler

func RegisterHandler(h Handler) {
	handlerRegistry = append(handlerRegistry, h)
}
func GetHandlers() []Handler {
	return handlerRegistry
}

var EmptyChecklist []byte
var EmptyChecklistItemsArray []*structs.ChecklistItem

type NavItem struct {
	Name string
	Path string
}

// Declaring the navbar
var NavList []NavItem = []NavItem{
	{
		Name: "Alle Einträge",
		Path: "/",
	},
	{
		Name: "Einträge löschen",
		Path: "/delete",
	},
	{
		Name: "Einträge zurücksetzen",
		Path: "/reset",
	},
	{ 
		Name: "Checkliste hinzufügen",
		Path: "/upload",
	},
}

func Nav(w http.ResponseWriter, r *http.Request){
  wd, err := os.Getwd()
  if err != nil{
    log.Fatal("couldn't get working directory: ", err)
  }
	var nav_tmpl = filepath.Join(wd, "static/nav.html")
	tmpl := template.Must(template.ParseFiles(nav_tmpl))

	err = tmpl.Execute(w, map[string]any{
		"Nav": UpdateNav(r),
	})
}


func GeneratePDF(w http.ResponseWriter, r *http.Request) {
		wd, err := os.Getwd()
		var static = filepath.Join(wd, "static")
		var print_tmpl = filepath.Join(static, "print.html")
		tmpl, err := template.ParseFiles(print_tmpl)
		var buf bytes.Buffer
		err = tmpl.Execute(&buf, map[string]any{
		})
		bodyBytes, err := io.ReadAll(&buf)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close() // Close the body after reading

		var pdfName string	
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

func DisplayChecklist(w http.ResponseWriter, r *http.Request){
  path := mux.Vars(r)["id"]
	log.Println(path)
  wd, err := os.Getwd()
  if err != nil{
    log.Fatal("couldn't get working directory: ", err)
  }
	var static = filepath.Join(wd, "static")
	var checklist_tmpl = filepath.Join(static, "checklist.html")
  tmpl := template.Must(template.ParseFiles(checklist_tmpl))
  err = tmpl.Execute(w, map[string]any{
  })
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    log.Fatal("", err)
  }
}


func DisplayUpload(w http.ResponseWriter, r *http.Request){
  wd, err := os.Getwd()
  if err != nil{
    log.Fatal("couldn't get working directory: ", err)
  }
	var static = filepath.Join(wd, "static")
	var new_tmpl = filepath.Join(static, "upload.html")
	var nav_tmpl = filepath.Join(static, "nav.html")
	var template_tmpl = filepath.Join(static, "template.html")

  tmpl := template.Must(template.ParseFiles(new_tmpl, nav_tmpl, template_tmpl))
	db := database.Init()
	allTemplates := database.GetAllTemplates(db)
  err = tmpl.Execute(w, map[string]any{
    "Nav" : UpdateNav(r),
		"Templates": allTemplates,
  })

  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    log.Fatal("", err)
  }

}
func ReceiveUpload(w http.ResponseWriter, r *http.Request){
	r.ParseMultipartForm(1 << 20)
	var buf bytes.Buffer
	file, header, err := r.FormFile("yaml")
	if err != nil {
		panic(err)
	}

	fields_raw := r.FormValue("fields_csv")
	desc_raw := r.FormValue("desc_csv")

	template_name := strings.Split(header.Filename, ".")[0]
	io.Copy(&buf, file)
	file_contents := buf.String()

	var result any
	err = yaml.Unmarshal([]byte(file_contents), &result)
	if err != nil {
		log.Fatalf("Error while validating the yaml in %s: %q\n", header.Filename, err)
	}

	fields_parsed := csv.NewReader(strings.NewReader(fields_raw))
	desc_parsed := csv.NewReader(strings.NewReader(desc_raw))

	fields, err := fields_parsed.Read()
	if err != nil {
		log.Fatalf("Error while reading parsed custom fields. %q \n", err)
	}
	desc, err := desc_parsed.Read()
	if err != nil {
		log.Fatalf("Error while reading parsed custom descriptions: %q \n", err)
	}

	if len(fields) == len(desc){
		db := database.Init()
		database.NewChecklistTemplate(db, template_name, file_contents, fields, desc)
	}
	http.Redirect(w, r, "/upload", http.StatusSeeOther)
}

type DescValueView struct {
	Desc string
	Value string
}

type EntriesView struct{
	Path string
	Data []DescValueView
}


// Returns description and value instead of the database-column and value.
func BuildEntriesView(custom_fields []database.CustomField, entries []database.ChecklistEntry) []EntriesView{
	var fieldsMap = make(map[string]string, len(custom_fields))
	for _, field := range custom_fields{
		fieldsMap[field.Key] = field.Desc
	}
	var result []EntriesView	
	for _, entry := range entries{
		var dataMap map[string]string
		err := json.Unmarshal([]byte(entry.Data), &dataMap)
		if err != nil{
			log.Fatalf("Error while unmarshaling json.\n Error: %q \n", err)
		}
		var viewMap []DescValueView
		for k, v := range dataMap {
				desc := fieldsMap[k]
				viewMap = append(viewMap, DescValueView{Desc: desc, Value: v})	
		}
		// format unix-time string to human-readable format
		viewMap = append(viewMap, DescValueView{
			Desc: "Erstellungsdatum",
			Value: time.Unix(entry.Date,0).Format("02-01-2006 15:04:05"),
		})
		result = append(result, EntriesView{
			Path: entry.Path,
			Data: viewMap,
		})
	}
	return result
}

// I want to save the current state of the active template when switching paths.
// So I add the Query to NavList.Path
func UpdateNav(r *http.Request)[]NavItem{
	update := make([]NavItem, len(NavList))
	copy(update, NavList)
	for i := range update{
		update[i].Path += "?"
		update[i].Path += r.URL.Query().Encode()
	}
	return update
}

// Takes './internal/handlers' as base-path.
// All paths 
func LoadTemplates(paths []string) *template.Template{
	wd, err := os.Getwd()
  if err != nil{
    log.Fatal("couldn't get working directory: ", err)
  }
	base := filepath.Join(wd, "internal", "handlers")
	log.Println(base)
	var full = make([]string, len(paths))
	for i, p := range paths{
		full[i] = filepath.Join(base,p)
	}
	return template.Must(template.ParseFiles(full...))
}

