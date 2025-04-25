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

	wkhtml "github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"github.com/gorilla/mux"
	"github.com/sqids/sqids-go"
	"gopkg.in/yaml.v3"
)

var EmptyChecklist []byte
var EmptyChecklistItemsArray []*structs.ChecklistItem

var NavList []structs.NavItem = []structs.NavItem{
	{
		Name: "Alle Einträge",
		Path: "/checklist",
	},
	{
		Name: "Einträge löschen",
		Path: "/checklist/delete",
	},
	{
		Name: "Einträge zurücksetzen",
		Path: "/checklist/reset",
	},
	{ 
		Name: "Checkliste hinzufügen",
		Path: "/checklist/upload",
	},
}

// Displays a form a new checklist-entry
// and a list with all previous entrys
func Home(w http.ResponseWriter, r *http.Request){
	wd, err := os.Getwd()
  if err != nil{
    log.Fatal("couldn't get working directory: ", err)
  }
	var static = filepath.Join(wd, "static")
	var new_tmpl = filepath.Join(static, "home.html")
	var nav_tmpl = filepath.Join(static, "nav.html")
	var select_tmpl = filepath.Join(static, "select.html")

  tmpl := template.Must(template.ParseFiles(new_tmpl, nav_tmpl, select_tmpl))

	db := database.Init()
	allTemplates := database.GetAllTemplates(db)
	// Set the URL with ?template=<template> if not already set
	if !r.URL.Query().Has("template") && len(allTemplates) > 0{
		u, err := url.Parse(r.URL.String())
		if err != nil {
			log.Fatalln("Error parsing GET-Request while loading ''.")	
		}
		q := u.Query()
		q.Set("template", allTemplates[0].Name)
		u.RawQuery = q.Encode()
		http.Redirect(w,r, u.String(), http.StatusFound)
		return
	}
  err = tmpl.Execute(w, map[string]any{
    "Nav" : NavList,
		"Templates": allTemplates,
		"Template_name": r.URL.Query().Get("template"),
  })

  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    log.Fatal("", err)
  }
}

func Options(w http.ResponseWriter, r *http.Request){

	template_name := r.URL.Query().Get("template")
	db := database.Init()
	custom_fields := database.GetAllFieldsForChecklist(db, template_name)
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
	custom_fields := database.GetAllFieldsForChecklist(db, template_name)
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
	query := r.URL.Query().Encode()
  wd, err := os.Getwd()
  if err != nil{
    log.Fatal("couldn't get working directory: ", err)
  }
	var nav_tmpl = filepath.Join(wd, "static/nav.html")
	tmpl := template.Must(template.ParseFiles(nav_tmpl))

	err = tmpl.Execute(w, map[string]any{
		"Nav": NavList,
		"Template": "?" + query,
	})
}


func NewEntry(w http.ResponseWriter, r *http.Request){
	template_name := r.FormValue("template")
	db := database.Init()
	template := database.GetChecklistTemplateByName(db, template_name)
	cols := database.GetAllFieldsForChecklist(db,template_name)
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



// Maybe this function is unnecessary,
// and I can call GET from the button
func RedirectToDownload(w http.ResponseWriter, r *http.Request){
  path :=  mux.Vars(r)["id"]
  // A redirect does not open a new windows with a pdf
  // so I need to do this hacky stuff with js
	cmd := fmt.Sprintf("<script>window.open('/checklist/download_%s', '_blank');</script>", path)
  fmt.Fprintf(w, cmd)
}

func GeneratePDF(w http.ResponseWriter, r *http.Request) {
		wd, err := os.Getwd()
		var static = filepath.Join(wd, "static")
		var print_tmpl = filepath.Join(static, "print.html")

		tmpl, err := template.ParseFiles(print_tmpl)

		// Generate html body into buffer
		var buf bytes.Buffer
		err = tmpl.Execute(&buf, map[string]any{
		})

		bodyBytes, err := io.ReadAll(&buf) 
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close() // Close the body after reading

		body := strings.NewReader(string(bodyBytes))

		pdfg, err :=  wkhtml.NewPDFGenerator()
		if err != nil{
			log.Fatalf("problem with pdf generator: %q", err)
			return
		}
    pdfg.PageSize.Set(wkhtml.PageSizeLetter)
    
    page := wkhtml.NewPageReader(body)
    page.Zoom.Set(0.95)

		pdfg.AddPage(page)

		err = pdfg.Create()
		if err != nil {
						log.Println(err)
						http.Error(w, "PDF creation error", http.StatusInternalServerError)
						return
		}
    var pdfName string 
		err = pdfg.WriteFile(pdfName)
		if err != nil {
						http.Error(w, "Failed to write PDF to file", http.StatusInternalServerError)
						return
		}

    file, err := os.Open(pdfName)
    if err != nil {
        http.Error(w, "File not found", http.StatusNotFound)
        return
    }
    defer file.Close()

    w.Header().Set("Content-Type", "application/pdf")
		disposition := fmt.Sprintf("attachment; filename=%s", pdfName)
    w.Header().Set("Content-Disposition", disposition)

    _, err = io.Copy(w, file)
    if err != nil {
        http.Error(w, "Error sending file", http.StatusInternalServerError)
        return
    }
    os.Remove(pdfName)
}

// Displays a page where existing database
// entrys can be altered
func DisplayDelete(w http.ResponseWriter, r *http.Request){
  wd, err := os.Getwd()
  if err != nil{
    log.Fatal("couldn't get working directory: ", err)
  }
	var static = filepath.Join(wd, "static")
	var new_tmpl = filepath.Join(static, "delete.html")
	var nav_tmpl = filepath.Join(static, "nav.html")

  tmpl := template.Must(template.ParseFiles(new_tmpl, nav_tmpl))

  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    log.Fatal("error parsing home template: ", err)
  }

  err = tmpl.Execute(w, map[string]any{
    "Nav" : NavList,
  })
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    log.Fatal("", err)
  }
}

// Handle a POST-Request to a path
func DeleteEntry(w http.ResponseWriter, r *http.Request){
  http.Redirect(w, r, "/checklist/delete", http.StatusSeeOther)
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
		"Nav": NavList,
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
  http.Redirect(w, r, "/checklist/reset", http.StatusSeeOther)
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
    "Nav" : NavList,
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
	http.Redirect(w, r, "/checklist/upload", http.StatusSeeOther)
}
