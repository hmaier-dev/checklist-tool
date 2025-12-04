package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"time"

	"github.com/hmaier-dev/checklist-tool/internal/database"
	"github.com/hmaier-dev/checklist-tool/internal/server"
)

// Just reads the database and displays entries
type DisplayHandler interface {
	New(srv *server.Server) // interface func to pass server values
	Routes() // sets routes to *mux.Router
	Display(w http.ResponseWriter, r *http.Request)
}

// Displays data but also writes to the database
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

// Contains data for rendering a single row in html
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
func BuildEntriesViewForTemplate(custom_fields []database.CustomField, entries []database.Entry) []EntryView {
	var result []EntryView
	for _, entry := range entries {
		result = append(result, BuildEntryViewForTemplate(custom_fields, &entry))
	}
	return result
}

// Connects the description of a column with it's value for a single entry.
// The DescValueView could also be build by a JOIN() in SQL.
// Maybe refactor it in the function, to reduce the codebase.
func BuildEntryViewForTemplate(custom_fields []database.CustomField, entry *database.Entry) EntryView {
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
	var t time.Time
	if entry.Date.Valid{
		t = time.Unix(entry.Date.Int64,0)		
	}else{
		t = time.Time{}
	}
	// format unix-time string to human-readable format
	viewMap = append(viewMap, DescValueView{
		Desc:  "Erstellungsdatum",
		Value: t.Format("02.01.2006 15:04:05"),
	})
	return EntryView{
		Path: entry.Path,
		Data: viewMap,
	}
}

// This is just used in delete.go
// What is the difference to the other EntryView returning function?
func ViewForEntry(db *sql.DB, ctx context.Context, entry database.GetAllEntriesPlusTemplateNameRow) EntryView {
	var dataMap map[string]string
	err := json.Unmarshal([]byte(entry.Data), &dataMap)
	if err != nil {
		log.Fatalf("Error while unmarshaling json.\n Error: %q \n", err)
	}
	var length int = len(dataMap)
	var viewMap []DescValueView = make([]DescValueView, length)
	q := database.New(db)
	custom_fields, err := q.GetCustomFieldsByTemplateName(ctx, entry.TemplateName)
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
	var t time.Time
	if entry.Date.Valid{
		t = time.Unix(entry.Date.Int64, 0)		
	}else{
		t = time.Time{}
	}
	return EntryView{
		TemplateName: entry.TemplateName,
		Date:         t.Format("02.01.2006 15:04:05"),
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
		"last": func(x int, a any) bool {
				return x == reflect.ValueOf(a).Len() - 1
		},
	}
	// add funcMap to base-template
	first := filepath.Base(full[0])
	tmpl := template.New(first).Funcs(template.FuncMap(funcMap))
	tmpl = template.Must(tmpl.ParseFiles(full...))
	return tmpl
}
