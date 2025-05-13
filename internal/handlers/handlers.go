package handlers

import (
	"database/sql"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
	"github.com/hmaier-dev/checklist-tool/internal/database"
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
		Name: "Neue Einträge",
		Path: "/",
	},
	{
		Name: "Alle Einträge",
		Path: "/all",
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

type EntryView struct{
	TemplateName string
	Date string
	Path string
	Data []DescValueView
}

type DescValueView struct {
	Desc string
	Value string
	Key string
}

// Connects the description of a column with it's value for an array of entries.
func BuildEntriesViewForTemplate(custom_fields []database.CustomField, entries []database.ChecklistEntry) []EntryView{
	var result []EntryView	
	for _, entry := range entries{
		result = append(result, BuildEntryViewForTemplate(custom_fields, entry))
	}
	return result
}
// Connects the description of a column with it's value for a single entry.
// The DescValueView could also be build by a JOIN() in SQL.
// Maybe refactor it in the function, to reduce the codebase.
func BuildEntryViewForTemplate(custom_fields []database.CustomField, entry database.ChecklistEntry) EntryView{
		var fieldsMap = make(map[string]string, len(custom_fields))
		for _, field := range custom_fields{
			fieldsMap[field.Key] = field.Desc
		}
		var dataMap map[string]string
		err := json.Unmarshal([]byte(entry.Data), &dataMap)
		if err != nil{
			log.Fatalf("Error while unmarshaling json.\n Error: %q \n", err)
		}
		var viewMap []DescValueView
		for _, field := range custom_fields{
			if val, ok := dataMap[field.Key];ok{
				viewMap = append(viewMap, DescValueView{
					Desc: field.Desc, 
					Value: val, 
					Key: field.Key,
				})
			}else{
				log.Fatalf("Key '%s' not found in dataMap: '%q' \n",field.Key,dataMap)
			}
		}
		// format unix-time string to human-readable format
		viewMap = append(viewMap, DescValueView{
			Desc: "Erstellungsdatum",
			Value: time.Unix(entry.Date,0).Format("02-01-2006 15:04:05"),
		})
		return EntryView{
			Path: entry.Path,
			Data: viewMap,			
		}
}

func ViewForEntry(db *sql.DB, entry database.EntryPlusChecklistName) EntryView{
		var dataMap map[string]string
		err := json.Unmarshal([]byte(entry.Data), &dataMap)
		if err != nil{
			log.Fatalf("Error while unmarshaling json.\n Error: %q \n", err)
		}
		var length int = len(dataMap)
		var viewMap []DescValueView = make([]DescValueView, length)
		custom_fields := database.GetAllCustomFieldsForTemplate(db, entry.TemplateName)
		var count int = 0
		// Use the order from the database-table 'custom_fields'
		for _, field := range custom_fields{
			if val, ok := dataMap[field.Key];ok{
				viewMap[count] = DescValueView{
					Desc: field.Desc, 
					Value: val, 
					Key: field.Key,
				}
			}else{
				log.Fatalf("Key '%s' not found in dataMap: '%q' \n",field.Key,dataMap)
			}
			count += 1
		}
		return EntryView{
			TemplateName: entry.TemplateName,
			Date: time.Unix(entry.Date,0).Format("02.01.2006 15:04:05"),
			Path: entry.Path,
			Data: viewMap,
		}
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

