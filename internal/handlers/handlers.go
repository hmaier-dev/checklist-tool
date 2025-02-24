package handlers

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
  "encoding/json"
  "github.com/hmaier-dev/checklist-tool/internal/checklist"
  "github.com/hmaier-dev/checklist-tool/internal/database"

	"github.com/gorilla/mux"
)


func CheckList(w http.ResponseWriter, r *http.Request){
  wd, err := os.Getwd()
  if err != nil{
    log.Fatal("couldn't get working directory: ", err)
  }
	var static = filepath.Join(wd, "static")
	var base = filepath.Join(static, "base.html")
  tmpl, err := template.ParseFiles(base)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    log.Fatal("error parsing the base template: ", err)
  }
  err = tmpl.Execute(w, nil) // write response to w
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    log.Fatal("", err)
  }

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

  tmpl, err := template.ParseFiles(new_tmpl)

  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    log.Fatal("error parsing home template: ", err)
  }

  err = tmpl.Execute(w, tmpl) // write response to w
  tmpl.ExecuteTemplate(w, "home", nil)
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

  if database.CheckIMEI(db,id) == false{
    http.Redirect(w, r, "/checklist/new", http.StatusSeeOther)
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

  // Print the unmarshaled JSON
  jsonData, err := json.MarshalIndent(items, "", "  ")
  if err != nil {
          log.Println("JSON marshal error:", err)
          return
  }
  fmt.Println(string(jsonData))

  err = tmpl.Execute(w, map[string]interface{}{
    "Items": items,
  })

  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    log.Fatal("", err)
  }
}

func Update(w http.ResponseWriter, r *http.Request){
  fmt.Println("Updating...")
  id := mux.Vars(r)["id"]
  fmt.Printf("Update Checklist for: %s \n", id)
}

