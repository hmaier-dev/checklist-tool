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
	"net/url"

	"github.com/hmaier-dev/checklist-tool/internal/database"
	"github.com/hmaier-dev/checklist-tool/internal/structs"
	"github.com/hmaier-dev/checklist-tool/internal/pdf"

	"github.com/sqids/sqids-go"
	"gopkg.in/yaml.v3"
	"github.com/gorilla/mux"
)

var EmptyChecklist []byte
var EmptyChecklistItemsArray []*structs.ChecklistItem

var NavList []structs.NavItem = []structs.NavItem{
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

// Displays a form a new checklist-entry
// and a list with all previous entrys
func Home(w http.ResponseWriter, r *http.Request){
	wd, err := os.Getwd()
  if err != nil{
    log.Fatal("couldn't get working directory: ", err)
  }
	var templates = []string{
		"home.html",
		"nav.html",
		"options.html",
		"entries.html",
	}
	var full = make([]string,len(templates))
	var static = filepath.Join(wd, "static")
	for i, t := range templates{
		full[i] = filepath.Join(static,t)
	}
  tmpl := template.Must(template.ParseFiles(full...))
	// This needs to be called here, to set ?template=
	// to the first template if none is set.
	db := database.Init()
	all := database.GetAllTemplates(db)
	active := r.URL.Query().Get("template")
	// TODO: turn this code if into a function
	// Set the URL with ?template=<template> if not already set
	if active == "" && len(all) > 0{
		u, err := url.Parse(r.URL.String())
		if err != nil {
			log.Fatalln("Error parsing GET-Request while loading ''.")	
		}
		q := u.Query()
		q.Set("template", all[0].Name)
		u.RawQuery = q.Encode()
		http.Redirect(w,r, u.String(), http.StatusFound)
		return
	}
	entries_raw := database.GetAllEntriesForChecklist(db, active)
	custom_fields := database.GetAllCustomFieldsForTemplate(db, active)
	entries_view := buildEntriesView(custom_fields,entries_raw)
	inputs := database.GetAllCustomFieldsForTemplate(db, active)
  err = tmpl.Execute(w, map[string]any{
		"Active": active,
    "Nav" : updateNav(r),
		"Templates": all,
		"Inputs": inputs,
		"Entries": entries_view,
  })

  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    log.Fatal("", err)
  }
}

func Options(w http.ResponseWriter, r *http.Request){
	template_name := r.URL.Query().Get("template")
	db := database.Init()
	custom_fields := database.GetAllCustomFieldsForTemplate(db, template_name)
  wd, err := os.Getwd()
  if err != nil{
    log.Fatal("couldn't get working directory: ", err)
  }
	var options = filepath.Join(wd, "static/options.html")
	tmpl := template.Must(template.ParseFiles(options))
	err = tmpl.Execute(w, map[string]any{
		"Inputs": custom_fields,
	})
}

func Entries(w http.ResponseWriter, r *http.Request){
	template_name := r.URL.Query().Get("template")
	db := database.Init()
	entries := database.GetAllEntriesForChecklist(db, template_name)
  wd, err := os.Getwd()
  if err != nil{
    log.Fatal("couldn't get working directory: ", err)
  }
	var entries_tmpl = filepath.Join(wd, "static/entries.html")
	tmpl := template.Must(template.ParseFiles(entries_tmpl))
	

	// building a map to access the descriptions by column names
	custom_fields := database.GetAllCustomFieldsForTemplate(db, template_name)
	var fieldsMap = make(map[string]string, len(custom_fields))
	for _, field := range custom_fields{
		fieldsMap[field.Key] = field.Desc
	}

	var result []structs.EntriesView	
	for _, entry := range entries{
		var dataMap map[string]string
		err := json.Unmarshal([]byte(entry.Data), &dataMap)
		if err != nil{
			log.Fatalf("Error while unmarshaling json.\n Error: %q \n", err)
		}
		var viewMap []structs.DescValueView
		for k, v := range dataMap {
				desc := fieldsMap[k]
				viewMap = append(viewMap, structs.DescValueView{Desc: desc, Value: v})	
		}
		// add creation date
		viewMap = append(viewMap, structs.DescValueView{
			Desc: "Erstellungsdatum",
			Value: time.Unix(entry.Date,0).Format("02-01-2006 15:04:05"),
		})
		result = append(result, structs.EntriesView{
			Path: entry.Path,
			Data: viewMap,
		})
	}
	err = tmpl.Execute(w, map[string]any{
		"Entries": result,
	})
}

func Nav(w http.ResponseWriter, r *http.Request){
  wd, err := os.Getwd()
  if err != nil{
    log.Fatal("couldn't get working directory: ", err)
  }
	var nav_tmpl = filepath.Join(wd, "static/nav.html")
	tmpl := template.Must(template.ParseFiles(nav_tmpl))

	err = tmpl.Execute(w, map[string]any{
		"Nav": updateNav(r),
	})
}


func NewEntry(w http.ResponseWriter, r *http.Request){
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
	path := GeneratePath(data)
	entry := structs.ChecklistEntry{
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

func GeneratePath(data map[string]string) string{
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

// Displays the page for deleting entries for the template passed in the GET-Request
func DisplayDelete(w http.ResponseWriter, r *http.Request){
  wd, err := os.Getwd()
  if err != nil{
    log.Fatal("couldn't get working directory: ", err)
  }
	var static = filepath.Join(wd, "static")
	var delete_tmpl = filepath.Join(static, "delete.html")
	var nav_tmpl = filepath.Join(static, "nav.html")

  tmpl := template.Must(template.ParseFiles(delete_tmpl, nav_tmpl))
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    log.Fatal("error parsing home template: ", err)
  }

	db := database.Init()
	all := database.GetAllTemplates(db)
  err = tmpl.Execute(w, map[string]any{
    "Nav" : updateNav(r),
		"Templates": all,
		"Active": r.URL.Query().Get("template"),
  })

  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    log.Fatalf("%q \n", err)
  }
}

// TODO: this code is reused from Entries
func DeleteableEntries(w http.ResponseWriter, r *http.Request){
	template_name := r.URL.Query().Get("template")
	db := database.Init()
	entries := database.GetAllEntriesForChecklist(db, template_name)
  wd, err := os.Getwd()
  if err != nil{
    log.Fatal("couldn't get working directory: ", err)
  }
	var entries_tmpl = filepath.Join(wd, "static/deleteEntries.html")
	tmpl := template.Must(template.ParseFiles(entries_tmpl))
	

	// building a map to access the descriptions by column names
	custom_fields := database.GetAllCustomFieldsForTemplate(db, template_name)
	var fieldsMap = make(map[string]string, len(custom_fields))
	for _, field := range custom_fields{
		fieldsMap[field.Key] = field.Desc
	}

	var result []structs.EntriesView	
	for _, entry := range entries{
		var dataMap map[string]string
		err := json.Unmarshal([]byte(entry.Data), &dataMap)
		if err != nil{
			log.Fatalf("Error while unmarshaling json.\n Error: %q \n", err)
		}
		var viewMap []structs.DescValueView
		for k, v := range dataMap {
				desc := fieldsMap[k]
				viewMap = append(viewMap, structs.DescValueView{Desc: desc, Value: v})	
		}
		// add creation date
		viewMap = append(viewMap, structs.DescValueView{
			Desc: "Erstellungsdatum",
			Value: time.Unix(entry.Date,0).Format("02-01-2006 15:04:05"),
		})
		result = append(result, structs.EntriesView{
			Path: entry.Path,
			Data: viewMap,
		})
	}
	err = tmpl.Execute(w, map[string]any{
		"Entries": result,
	})

}

// Handle a POST-Request to a path
func DeleteEntry(w http.ResponseWriter, r *http.Request){
  http.Redirect(w, r, "/delete", http.StatusSeeOther)
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


func DisplayReset(w http.ResponseWriter, r *http.Request){
  wd, err := os.Getwd()
  if err != nil{
    log.Fatal("couldn't get working directory: ", err)
  }
	var static = filepath.Join(wd, "static")
	var new_tmpl = filepath.Join(static, "reset.html")

	var nav_tmpl = filepath.Join(static, "nav.html")

  tmpl := template.Must(template.ParseFiles(new_tmpl, nav_tmpl))

  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    log.Fatal("error parsing home template: ", err)
  }

  err = tmpl.Execute(w, map[string]any{
		"Nav": updateNav(r),
  })
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    log.Fatal("", err)
  }

}

// Handle POST-Request for resetting a checklist
func ResetChecklistForEntry(w http.ResponseWriter, r *http.Request){
  db := database.Init()
  defer db.Close() // Make sure to close the database when done
  http.Redirect(w, r, "/reset", http.StatusSeeOther)
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
    "Nav" : updateNav(r),
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

// Helper function to structure []structs.ChecklistEntry into []structs.EntriesView.
// Instead of using this function I could access the ChecklistEntry.Data-Fields in the template like this
//
// {{ range $key, $value := .Data }}
// <td >{{ $key }}</td>
// {{ end }}
// {{ range $key, $value := .Data }}
// <td >{{ $value }}</td>
// {{ end }}
//
// For me it seems a lot cleaner to build a struct for this.
func buildEntriesView(custom_fields []structs.CustomFields, entries []structs.ChecklistEntry) []structs.EntriesView{
	var fieldsMap = make(map[string]string, len(custom_fields))
	for _, field := range custom_fields{
		fieldsMap[field.Key] = field.Desc
	}
	var result []structs.EntriesView	
	for _, entry := range entries{
		var dataMap map[string]string
		err := json.Unmarshal([]byte(entry.Data), &dataMap)
		if err != nil{
			log.Fatalf("Error while unmarshaling json.\n Error: %q \n", err)
		}
		var viewMap []structs.DescValueView
		for k, v := range dataMap {
				desc := fieldsMap[k]
				viewMap = append(viewMap, structs.DescValueView{Desc: desc, Value: v})	
		}
		// format unix-time string to human-readable format
		viewMap = append(viewMap, structs.DescValueView{
			Desc: "Erstellungsdatum",
			Value: time.Unix(entry.Date,0).Format("02-01-2006 15:04:05"),
		})
		result = append(result, structs.EntriesView{
			Path: entry.Path,
			Data: viewMap,
		})
	}
	return result
}

// I want to save the current state of the active template when switching paths.
// So I add the Query to NavList.Path
func updateNav(r *http.Request)[]structs.NavItem{
	update := make([]structs.NavItem, len(NavList))
	copy(update, NavList)
	for i := range update{
		update[i].Path += "?"
		update[i].Path += r.URL.Query().Encode()
	}
	return update
}

