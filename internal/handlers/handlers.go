package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/hmaier-dev/checklist-tool/internal/database"
)

// Just sets the routes and displays html
type DisplayHandler interface {
	Routes(router *mux.Router)
	Display(w http.ResponseWriter, r *http.Request)
}

// Writes
type ActionHandler interface {
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
		Name: "Checklisten verwalten",
		Path: "/upload",
	},
}

func Nav(w http.ResponseWriter, r *http.Request) {
	tmpl := LoadTemplates([]string{"nav.html"})
	err := tmpl.Execute(w, map[string]any{
		"Nav": NavList,
	})
	if err != nil {
		log.Fatalf("Something went wrong executing the 'nav.html' template.\n %q \n", err)
	}
}

// By reading the header of the GET-request, we get the path of the currentPage
// so can get created or get appended.
// Is called from within 'history.html'
func HistoryBreadcrumb(w http.ResponseWriter, r *http.Request) {
	localStorage := r.FormValue("lastPages")
	var lastPages []string
	err := json.Unmarshal([]byte(localStorage),&lastPages)
	if err != nil{
		log.Fatalf("%q", err)
	}
	// To build the TabDescription for each breadcrumb, we need the database
	db := database.Init()
	var entries []*database.ChecklistEntry
	for _, path := range lastPages{
		e, err	:= database.GetEntryByPath(db, path)
		if err != nil {
			log.Printf("The path '%s' is not existent? We are skipping it.", path )
		}else{
			entries = append(entries, e)
		}
	}
	var history = []struct {
		// Is the path to navigate to
		Path           string
		// Actual content of the breadcrumb
		TabDescription string
	}{}
	// The values of a schema are organized in the table `tab_desc_schema`. 
	// We access them by template_id (which is the primary key for all checklist metadata).
	for _, entry := range entries{
		complete_schema := database.GetTabDescriptionsByID(db, entry.Template_id)
		// The schema just have the keys, but we want the data which is in entry.Data
		var data map[string]string
		err := json.Unmarshal([]byte(entry.Data),&data)
		if err != nil{
			log.Fatalln("Unmarshaling json from db wen't wrong.")
			return
		}
		// This inner loop combines the different db-entries for the TabDescription
		var result string
		for i, t := range complete_schema{
			if i == len(complete_schema)-1 {
				result += data[t.Value]
			} else {
				result += data[t.Value] + " | "
			}
		}
		history = append(history, struct{Path string; TabDescription string}{Path: entry.Path, TabDescription: result})
	}

	tmpl := LoadTemplates([]string{"breadcrumb-history.html"})
	err = tmpl.Execute(w, map[string]any{
		"History": history,
	})
	if err != nil {
		log.Fatalf("Something went while building the breadcrumb history...\n %q \n", err)
	}
}

// Returns the history as marshaled json.
// history is ordered ascending (from new to old)
func History(w http.ResponseWriter, r *http.Request) {
	lastPages, err := appendHistory(r)
	if err != nil{
		log.Fatalf("Couldn't append the history.\n Error: %q", err)
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(lastPages); err != nil {
		http.Error(w, "failed to encode JSON", http.StatusInternalServerError)
		return
	}
}

// Returns []string in ascending order (from oldest to newest)
func appendHistory(r *http.Request) ([]string, error){
	currentUrl := r.Header.Get("Referer")
	u, err := url.Parse(currentUrl)
	if err != nil {
		return nil, fmt.Errorf("Something wen't wrong when parsing the URL received on /history...")
	}
	split := strings.Split(u.Path, "/")
	currentPath := split[len(split)-1]
	localStorage := r.FormValue("lastPages")
	// Set the currentPath as the oldest page
	if len(localStorage) == 0{
		return []string{currentPath}, nil
	}
	// from oldest to newest
	var history []string
	err = json.Unmarshal([]byte(localStorage),&history)
	// No entry should be duplicated
	if !slices.Contains(history, currentPath){
		history = append(history, currentPath)	
	}
	// We need a db connection to check whether pre-existent paths in the browsers storage, are still in the database.
	// If a path is present in the browser but not in database, building the breadcrumb would fail
	db := database.Init()
	var clean []string
	for _, p := range history{
		_, err:= database.GetEntryByPath(db, p)
		if err == nil{
			// This path was found in the database
			clean = append(clean, p)
		}
	}
	return clean, nil
}


type EntryView struct {
	TemplateName string
	Date         string
	Path         string
	Data         []DescValueView
}

type DescValueView struct {
	Desc  string
	Value string
	Key   string
}

// Connects the description of a column with it's value for an array of entries.
func BuildEntriesViewForTemplate(custom_fields []database.CustomField, entries []*database.ChecklistEntry) []EntryView {
	var result []EntryView
	for _, entry := range entries {
		result = append(result, BuildEntryViewForTemplate(custom_fields, entry))
	}
	return result
}

// Connects the description of a column with it's value for a single entry.
// The DescValueView could also be build by a JOIN() in SQL.
// Maybe refactor it in the function, to reduce the codebase.
func BuildEntryViewForTemplate(custom_fields []database.CustomField, entry *database.ChecklistEntry) EntryView {
	var fieldsMap = make(map[string]string, len(custom_fields))
	for _, field := range custom_fields {
		fieldsMap[field.Key] = field.Desc
	}
	var dataMap map[string]string
	err := json.Unmarshal([]byte(entry.Data), &dataMap)
	if err != nil {
		log.Fatalf("Error while unmarshaling json.\n Error: %q \n", err)
	}
	var viewMap []DescValueView
	for _, field := range custom_fields {
		if val, ok := dataMap[field.Key]; ok {
			viewMap = append(viewMap, DescValueView{
				Desc:  field.Desc,
				Value: val,
				Key:   field.Key,
			})
		} else {
			log.Fatalf("Key '%s' not found in dataMap: '%q' \n", field.Key, dataMap)
		}
	}
	// format unix-time string to human-readable format
	viewMap = append(viewMap, DescValueView{
		Desc:  "Erstellungsdatum",
		Value: time.Unix(entry.Date, 0).Format("02.01.2006 15:04:05"),
	})
	return EntryView{
		Path: entry.Path,
		Data: viewMap,
	}
}

func ViewForEntry(db *sql.DB, entry database.EntryPlusChecklistName) EntryView {
	var dataMap map[string]string
	err := json.Unmarshal([]byte(entry.Data), &dataMap)
	if err != nil {
		log.Fatalf("Error while unmarshaling json.\n Error: %q \n", err)
	}
	var length int = len(dataMap)
	var viewMap []DescValueView = make([]DescValueView, length)
	custom_fields := database.GetAllCustomFieldsForTemplate(db, entry.TemplateName)
	var count int = 0
	// Use the order from the database-table 'custom_fields'
	for _, field := range custom_fields {
		if val, ok := dataMap[field.Key]; ok {
			viewMap[count] = DescValueView{
				Desc:  field.Desc,
				Value: val,
				Key:   field.Key,
			}
		} else {
			log.Fatalf("Key '%s' not found in dataMap: '%q' \n", field.Key, dataMap)
		}
		count += 1
	}
	return EntryView{
		TemplateName: entry.TemplateName,
		Date:         time.Unix(entry.Date, 0).Format("02.01.2006 15:04:05"),
		Path:         entry.Path,
		Data:         viewMap,
	}
}

// Takes './internal/handlers' as base-path.
// Keep in mind that paths[0] must be the base/root-template
// that uses all other templates!
func LoadTemplates(paths []string) *template.Template {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal("couldn't get working directory: ", err)
	}
	base := filepath.Join(wd, "internal", "handlers")
	var full = make([]string, len(paths))
	for i, p := range paths {
		full[i] = filepath.Join(base, p)
	}
	funcMap := template.FuncMap{
		"arr": func(item ...any) []any { return item },
	}
	// add funcMap to base-template
	first := filepath.Base(full[0])
	tmpl := template.New(first).Funcs(template.FuncMap(funcMap))
	tmpl = template.Must(tmpl.ParseFiles(full...))
	return tmpl
}
