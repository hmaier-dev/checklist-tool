package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
)

func HealthCheck(w http.ResponseWriter, r *http.Request){
  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(http.StatusOK)
  json.NewEncoder(w).Encode(map[string]string{"status": "ok"})

}
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

