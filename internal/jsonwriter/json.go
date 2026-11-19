package jsonwriter

import (
	"encoding/json"
	"net/http"
)


func Write(w http.ResponseWriter, status int, res any){
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(res)
}