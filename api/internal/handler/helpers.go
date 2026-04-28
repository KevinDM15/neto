// Package handler implementa los handlers HTTP de la API de Neto.
// Es la única capa que conoce net/http y Chi.
package handler

import (
	"encoding/json"
	"net/http"
)

// writeJSON escribe una respuesta JSON con el status code dado.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError escribe un error JSON con formato {"error":"mensaje"}.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// decode parsea el body JSON del request en v.
func decode(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}
