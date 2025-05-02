package handlers

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/hmaier-dev/checklist-tool/internal/database"
	"github.com/gorilla/mux"
)


// Just sets the routes and displays html
type DisplayHandler interface{
	Routes(router *mux.Router)
	Display(w http.ResponseWriter, r *http.Request)	
}

// Writes
type ActionHandler interface{
	DisplayHandler
	Entries(w http.ResponseWriter, r *http.Request)	
	Execute(w http.ResponseWriter, r *http.Request)
}

var handlerRegistry []DisplayHandler

func RegisterHandler(h DisplayHandler) {
	handlerRegistry = append(handlerRegistry, h)
}
func GetHandlers() []DisplayHandler {
	return handlerRegistry
}


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
	tmpl := LoadTemplates([]string{"nav.html"})
	err := tmpl.Execute(w, map[string]any{
		"Nav": UpdateNav(r),
	})
	if err != nil{
		log.Fatalf("Something went wrong executing the 'nav.html' template.\n %q \n", err)
	}
}

type DescValueView struct {
	Desc string
	Value string
}

type EntryView struct{
	Path string
	Data []DescValueView
}


// Returns description and value instead of the database-column and value.
func BuildEntriesView(custom_fields []database.CustomField, entries []database.ChecklistEntry) []EntryView{
	var fieldsMap = make(map[string]string, len(custom_fields))
	for _, field := range custom_fields{
		fieldsMap[field.Key] = field.Desc
	}
	var result []EntryView	
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
		result = append(result, EntryView{
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
// Keep in mind that paths[0] must be the base/root-template
// that uses all other templates! 
func LoadTemplates(paths []string) *template.Template{
	wd, err := os.Getwd()
  if err != nil{
    log.Fatal("couldn't get working directory: ", err)
  }
	base := filepath.Join(wd, "internal", "handlers")
	var full = make([]string, len(paths))
	for i, p := range paths{
		full[i] = filepath.Join(base,p)
	}
	funcMap := template.FuncMap{
		"arr": func (item ...any) []any { return item },
	}
	// add funcMap to base-template
	first := filepath.Base(full[0])
	tmpl := template.New(first).Funcs(template.FuncMap(funcMap))
	tmpl = template.Must(tmpl.ParseFiles(full...))
	return tmpl
}

