package upload

import (
	"bytes"
	"encoding/csv"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/hmaier-dev/checklist-tool/internal/database"
	"github.com/hmaier-dev/checklist-tool/internal/handlers"
	"gopkg.in/yaml.v3"
)

type UploadHandler struct{}

var _ handlers.DisplayHandler = (*UploadHandler)(nil)


// Sets / and all its subroutes
func (h *UploadHandler)	Routes(router *mux.Router){
	sub := router.PathPrefix("/upload").Subrouter()
  sub.HandleFunc("", h.Display).Methods("Get")
  sub.HandleFunc("", h.Execute).Methods("POST")
}

// Return html to http.ResponseWriter for /
func (h *UploadHandler) Display(w http.ResponseWriter, r *http.Request){
	var templates = []string{
		"upload/templates/upload.html",
		"nav.html",
		"upload/templates/template.html",
	}
	db := database.Init()
	all := database.GetAllTemplates(db)
	tmpl := handlers.LoadTemplates(templates)
	err := tmpl.Execute(w, map[string]any{
		"Templates": all,
		"Nav": handlers.UpdateNav(r),
  })
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    log.Fatal("", err)
  }

}

// Runs when submit-button on / is pressed
func (h *UploadHandler) Execute(w http.ResponseWriter, r *http.Request){
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

func init(){
	handlers.RegisterHandler(&UploadHandler{})
}
