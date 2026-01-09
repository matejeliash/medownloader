package server

import (
	"encoding/json"
	"net/http"
)

// decode data from resp body into any struct
func decodeJson(r *http.Request, data any) error {
	return json.NewDecoder(r.Body).Decode(data)
}

// encode any struct and write data as JSON
func encodeJson(w http.ResponseWriter, v any, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)

}

// encode error message and write it as JSON
func encodeErr(w http.ResponseWriter, errorMsg string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"err": errorMsg})
}
