package history

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/gorilla/mux"
	"github.com/hmaier-dev/checklist-tool/internal/database"
	"github.com/hmaier-dev/checklist-tool/internal/handlers"
	"github.com/hmaier-dev/checklist-tool/internal/server"
)

type HistoryHandler struct{
	Router *mux.Router
	DB *sql.DB
}

var _ handlers.DisplayHandler = (*HistoryHandler)(nil)

func (h *HistoryHandler) New(srv *server.Server){
	h.Router = srv.Router
	h.DB = srv.DB
}

func (h *HistoryHandler) Routes(){
	h.Router.HandleFunc("/history-breadcrumb", h.Display).Methods("POST")
	// Because the paths stored client-side are very long POST requests are used to handle them
  h.Router.HandleFunc("/history-data", h.History).Methods("POST")
}

// By reading the header of the GET-request, we get the path of the currentPage
// so can get created or get appended.
// Is called from within 'history.html'
func (h *HistoryHandler)	Display(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	localStorage := r.FormValue("lastPages")
	var lastPages []string
	err := json.Unmarshal([]byte(localStorage),&lastPages)
	if err != nil{
		http.Error(w,"Could not use unmarshal the list from localStorage.", http.StatusBadRequest)
		return
	}
	// To build the TabDescription for each breadcrumb, we need the database
	q := database.New(h.DB)
	var entries []*database.Entry
	for _, path := range lastPages{
		e, err	:= q.GetEntryByPath(ctx, path)
		if err != nil {
			log.Printf("The path '%s' is not existent? We are skipping it.", path )
		}else{
			entries = append(entries, &e)
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
		completeSchema, err := q.GetTabDescriptionsByTemplateID(ctx, entry.TemplateID)
		if err != nil{
			msg := fmt.Sprintf("Could not fetch TabDescription for, Data: '%s', Path: '%s'", entry.Data, entry.Path)
			log.Println(msg)
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}
		// The schema just have the keys, but we want the data which is in entry.Data
		var data map[string]string
		err = json.Unmarshal([]byte(entry.Data),&data)
		if err != nil{
			msg := fmt.Sprintf("Could not unmarshal '%s' from 'entries'-table.", entry.Data)
			log.Println(msg)
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}
		// This inner loop combines the values in entry.Data to the TabDescription
		var result string
		for i, t := range completeSchema{
			if i == len(completeSchema)-1 {
				result += data[t.Value]
			} else {
				result += data[t.Value] + " | "
			}
		}
		history = append(history, struct{Path string; TabDescription string}{Path: entry.Path, TabDescription: result})
	}
	slices.Reverse(history)
	tmpl := handlers.LoadTemplates([]string{"history/templates/breadcrumb-history.html"})

	err = tmpl.Execute(w, map[string]any{
		// do not display the newest
		"History": history[1:],
	})
	if err != nil {
		log.Fatalf("Something went while building the breadcrumb history...\n %q \n", err)
	}
}

// Returns the history as marshaled json.
// history is ordered ascending (from new to old)
func (h *HistoryHandler) History(w http.ResponseWriter, r *http.Request) {
	lastPages, err := h.appendHistory(r)
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
func (h *HistoryHandler) appendHistory(r *http.Request) ([]string, error){
	ctx :=  r.Context()
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
	// From oldest to newest
	var history []string
	err = json.Unmarshal([]byte(localStorage), &history)
	// Change history order, the newest must be last!
	for i, h := range history{
		if h == currentPath{
			history = append(history[:i], history[i+1:]...)
		}
	}
	history = append(history, currentPath)
	// We need a db connection to check whether pre-existent paths in the browsers storage, are still in the database.
	// If a path is present in the browser but not in database, building the breadcrumb would fail
	q := database.New(h.DB)
	var clean []string
	for _, p := range history{
		_, err:= q.GetEntryByPath(ctx, p)
		if err == nil{
			// This path was found in the database
			clean = append(clean, p)
		}
	}
	// Limit the list 
	if len(clean) > 9{
		// drop the oldest item
		return clean[1:], nil
	}
	return clean, nil
}

func init() {
	handlers.RegisterHandler(&HistoryHandler{})
}
