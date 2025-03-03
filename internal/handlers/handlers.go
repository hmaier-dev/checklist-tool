package handlers

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
  "time"

	"github.com/hmaier-dev/checklist-tool/internal/structs"
	"github.com/hmaier-dev/checklist-tool/internal/database"
	"github.com/hmaier-dev/checklist-tool/internal/helper"

	wkhtml "github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"github.com/gorilla/mux"

	"gopkg.in/yaml.v3"
)

var EmptyChecklist []byte
var EmptyChecklistItemsArray []*structs.ChecklistItem

// Displays a form a new checklist-entry
// and a list with all previous entrys
func Home(w http.ResponseWriter, r *http.Request){
  wd, err := os.Getwd()
  if err != nil{
    log.Fatal("couldn't get working directory: ", err)
  }
	var static = filepath.Join(wd, "static")
	var new_tmpl = filepath.Join(static, "home.html")

  tmpl, err := template.ParseFiles(new_tmpl)

  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    log.Fatal("error parsing home template: ", err)
  }

  db := database.Init()
  data, err := database.GetAllEntrysReversed(db)

  err = tmpl.Execute(w, map[string]interface{}{
    "Entries" : data,
  })
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    log.Fatal("", err)
  }
}

// POST-Endpoint to receive the request made by the formular
func NewEntry(w http.ResponseWriter, r *http.Request){
  if r.Method != http.MethodPost {
    http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
    return
	}
  form := structs.FormularData{
    IMEI : r.FormValue("imei"),
    ITA : r.FormValue("ita"),
    Name: r.FormValue("name"),
    Ticket: r.FormValue("ticket"),
    Model: r.FormValue("model"),
  }

  db := database.Init()
  defer db.Close() // Make sure to close the database when done

  database.NewEntry(db, form)
  
  redirectTo := fmt.Sprintf("/checklist/%s", form.IMEI)
  http.Redirect(w, r, redirectTo, http.StatusSeeOther)
}

// Based on the IMEI a fitting db-entry will get loaded
func Display(w http.ResponseWriter, r *http.Request){
  id := mux.Vars(r)["id"]
  db := database.Init()

  if database.IMEIalreadyExists(db,id) == false{
    http.Redirect(w, r, "/checklist", http.StatusSeeOther)
    return
  }

  data, err := database.GetDataByIMEI(db, id)
  if err != nil {
    http.Error(w, "Database error", http.StatusInternalServerError)
    log.Println("Database error :", err)
    return
  }

  var items []*structs.ChecklistItem
	err = yaml.Unmarshal([]byte(data.Yaml), &items)
	if err != nil {
		fmt.Println("Error parsing YAML:", err)
		return
	}

  helper.AddDataToEveryEntry(data.IMEI, items)

  if err != nil {
      http.Error(w, "Invalid JSON", http.StatusInternalServerError)
      log.Println("JSON unmarshal error: ", err)
      return
  }

  wd, err := os.Getwd()
  if err != nil{
    log.Fatal("couldn't get working directory: ", err)
  }

	var static = filepath.Join(wd, "static")
	var checklist = filepath.Join(static, "checklist.html")

  tmpl, err := template.ParseFiles(checklist)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    log.Fatal("error parsing base and new template: ", err)
  }


  info := struct{
    IMEI string
    ITA string
    Name string
    Ticket  string
    Model string
  }{
    IMEI: data.IMEI,
    ITA: data.ITA,
    Name: data.Name,
    Ticket: data.Ticket, 
    Model: data.Model, 
  }

  err = tmpl.Execute(w, map[string]interface{}{
    "Items": items,
    "Info": info,
  })

  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    log.Fatal("", err)
  }
}

// Show the blanko checklist
func DisplayBlanko(w http.ResponseWriter, r *http.Request){
  var items []*structs.ChecklistItem
  err := yaml.Unmarshal([]byte(EmptyChecklist), &items)
  if err != nil {
      http.Error(w, "Invalid Yaml", http.StatusInternalServerError)
      log.Println("Yaml unmarshal error: ", err)
      return
  }

  wd, err := os.Getwd()
  if err != nil{
    log.Fatal("couldn't get working directory: ", err)
  }

	var static = filepath.Join(wd, "static")
	var checklist = filepath.Join(static, "checklist.html")

  tmpl, err := template.ParseFiles(checklist)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    log.Fatal("error parsing base and new template: ", err)
  }

  err = tmpl.Execute(w, map[string]interface{}{
    "Items": items,
  })

  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    log.Fatal("", err)
  }

}

func Update(w http.ResponseWriter, r *http.Request){

  // Parse form
  imei :=  mux.Vars(r)["id"]
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
  alteredItem := structs.ChecklistItem{
    Task: r.Form.Get("task"),
    Checked: checked,
  }
  
  // Fetch Row from Database
  db := database.Init()
  row, err := database.GetDataByIMEI(db, imei)
  if err != nil{
    log.Fatalf("error fetching data by imei: %q", err)
  }
  var oldItems []*structs.ChecklistItem
  err = yaml.Unmarshal([]byte(row.Yaml), &oldItems)

  helper.ChangeCheckedStatus(alteredItem, oldItems)

  yamlBytes, err := yaml.Marshal(oldItems)
  if err != nil {
      log.Println("Error marshaling Yaml: ", err)
      return
  }
  // Submit Altered Row to database
  database.UpdateYamlByIMEI(db, imei, string(yamlBytes))

}

// Maybe this function is unnecessary,
// and I can call GET from the button
func RedirectToDownload(w http.ResponseWriter, r *http.Request){
  imei :=  mux.Vars(r)["id"]
  // A redirect does not open a new windows with a pdf
  // so I need to do this hacky stuff with js
	cmd := fmt.Sprintf("<script>window.open('/checklist/download_%s', '_blank');</script>", imei)
  fmt.Fprintf(w, cmd)
}

func GeneratePDF(w http.ResponseWriter, r *http.Request) {
		imei :=  mux.Vars(r)["id"]

		wd, err := os.Getwd()
		var static = filepath.Join(wd, "static")
		var print_tmpl = filepath.Join(static, "print.html")

		tmpl, err := template.ParseFiles(print_tmpl)

		db := database.Init()
		row, err := database.GetDataByIMEI(db, imei)

		var items []*structs.ChecklistItem
		err = yaml.Unmarshal([]byte(row.Yaml), &items)

		// What is this??? can't I just give row to Info??
		info := struct{
			IMEI string
			ITA string
			Name string
			Ticket  string
			Model string
		}{
			IMEI: row.IMEI,
			ITA: row.ITA,
			Name: row.Name,
			Ticket: row.Ticket, 
			Model: row.Model, 
		}

		now := time.Now()
    curr_date := now.Format("02.01.2006, 15:04:05")

		// Generate html body into buffer
		var buf bytes.Buffer
		err = tmpl.Execute(&buf, map[string]any{
			"Items": items,
			"Info": info,
      "Date": curr_date,
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
		fDate := now.Format("20060102")
		parts := strings.Fields(info.Name)
    var pdfName string 
    if len(parts) > 1 {
      pdfName = fmt.Sprintf("%s_%s_%s_%s_%s.pdf",fDate,parts[0],parts[1],info.Model,info.IMEI)
    }else{
      pdfName = fmt.Sprintf("%s_%s_%s_%s.pdf",fDate,parts[0],info.Model,info.IMEI)
    }

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
