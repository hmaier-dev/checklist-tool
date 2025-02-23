package handlers

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
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

func Formular(w http.ResponseWriter, r *http.Request){
  wd, err := os.Getwd()
  if err != nil{
    log.Fatal("couldn't get working directory: ", err)
  }
	var static = filepath.Join(wd, "static")
	var base = filepath.Join(static, "base.html")
	var new_tmpl = filepath.Join(static, "new.html")

  tmpl, err := template.ParseFiles(base, new_tmpl)

  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    log.Fatal("error parsing base and new template: ", err)
  }

  err = tmpl.Execute(w, tmpl) // write response to w
  tmpl.ExecuteTemplate(w, "base", nil)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    log.Fatal("", err)
  }
}

func NewEntry(w http.ResponseWriter, r *http.Request){
  if r.Method != http.MethodPost {
    http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
    return
	}
  form := checklist.FormularData{
    Imei : r.FormValue("imei"),
    Name: r.FormValue("name"),
    Ticket: r.FormValue("ticket"),
    Modell: r.FormValue("model"),
  }

  db := database.Init()
  defer db.Close() // Make sure to close the database when done

  database.NewEntry(form)

  http.Redirect(w, r, "/checklist/new", http.StatusSeeOther)
}


func Display(w http.ResponseWriter, r *http.Request){
  fmt.Println("Displaying...")
  id := mux.Vars(r)["id"]
  fmt.Printf("Display Checklist IMEI: %s\n", id)

}

func Update(w http.ResponseWriter, r *http.Request){
  fmt.Println("Updating...")
  id := mux.Vars(r)["id"]
  fmt.Printf("Update Checklist for: %s \n", id)
}

