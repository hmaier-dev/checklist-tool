package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/hmaier-dev/checklist-tool/internal/checklist"
	"github.com/hmaier-dev/checklist-tool/internal/database"
	"github.com/hmaier-dev/checklist-tool/internal/helper"

	wkhtml "github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"github.com/gorilla/mux"
)

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
  form := checklist.FormularData{
    IMEI : r.FormValue("imei"),
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

  var items []*checklist.ChecklistItem
  err = json.Unmarshal([]byte(data.Json), &items)

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
    Name string
    Ticket  string
    Model string
  }{
    IMEI: data.IMEI,
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



func DisplayBlanko(w http.ResponseWriter, r *http.Request){
  filename := "./test_checklist.json"
  blankoChecklist, err := os.ReadFile(filename)
  if err != nil {
		log.Fatal("Problem reading the empty json:", err)
  }

  var items []*checklist.ChecklistItem
  err = json.Unmarshal([]byte(blankoChecklist), &items)

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
  alteredItem := checklist.ChecklistItem{
    Task: r.Form.Get("task"),
    Checked: checked,
  }
  
  // Fetch Row from Database
  db := database.Init()
  row, err := database.GetDataByIMEI(db, imei)
  if err != nil{
    log.Fatalf("error fetching data by imei: %q", err)
  }
  var oldItems []*checklist.ChecklistItem
  err = json.Unmarshal([]byte(row.Json), &oldItems)

  helper.ChangeCheckedStatus(alteredItem, oldItems)

  jsonBytes, err := json.Marshal(oldItems)
  if err != nil {
      log.Println("Error marshaling JSON:", err)
      return
  }
  // Submit Altered Row to database
  database.UpdateJsonByIMEI(db, imei, string(jsonBytes))

}

func GeneratePDF(w http.ResponseWriter, r *http.Request){
  imei :=  mux.Vars(r)["id"]
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

  wd, err := os.Getwd()
	var static = filepath.Join(wd, "static")
	var print_tmpl = filepath.Join(static, "print.html")

  tmpl, err := template.ParseFiles(print_tmpl)

  db := database.Init()
  row, err := database.GetDataByIMEI(db, imei)

  var items []*checklist.ChecklistItem
  err = json.Unmarshal([]byte(row.Json), &items)

  info := struct{
    IMEI string
    Name string
    Ticket  string
    Model string
  }{
    IMEI: row.IMEI,
    Name: row.Name,
    Ticket: row.Ticket, 
    Model: row.Model, 
  }

  // Generate html body into buffer
  var buf bytes.Buffer
  err = tmpl.Execute(&buf, map[string]any{
    "Items": items,
    "Info": info,
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
  pdfg.AddPage(wkhtml.NewPageReader(body))

  err = pdfg.Create()
  if err != nil {
          log.Println(err)
          http.Error(w, "PDF creation error", http.StatusInternalServerError)
          return
  }

  // now := time.Now()
  // formattedDate := now.Format("20060102")
  pdfName := "test.pdf"
  err = pdfg.WriteFile(pdfName)
  if err != nil {
          http.Error(w, "Failed to write PDF to file", http.StatusInternalServerError)
          return
  }
  // A redirect does not open a new windows with a pdf
  // so I need to do this hacky stuff with js
  fmt.Fprintf(w, "<script>window.open('/checklist/serve-pdf', '_blank');</script>")
}

func ServeStaticPDF(w http.ResponseWriter, r *http.Request) {
    filePath := "test.pdf" // Path to your static PDF

    file, err := os.Open(filePath)
    if err != nil {
        http.Error(w, "File not found", http.StatusNotFound)
        return
    }
    defer file.Close()

    w.Header().Set("Content-Type", "application/pdf")
    w.Header().Set("Content-Disposition", "attachment; filename=test.pdf")

    _, err = io.Copy(w, file)
    if err != nil {
        http.Error(w, "Error sending file", http.StatusInternalServerError)
        return
    }
}
