package handlers

import(
  "net/http"
  "encoding/json"
)

func HealthCheck(w http.ResponseWriter, r *http.Request){
  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(http.StatusOK)
  json.NewEncoder(w).Encode(map[string]string{"status": "ok"})

}
