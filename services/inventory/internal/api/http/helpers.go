package httpapi

import (
	"encoding/json"
	"net/http"
)

// writeJSON сериализует v в JSON и пишет в ResponseWriter.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError пишет стандартный JSON-ответ с ошибкой.
func writeError(w http.ResponseWriter, status int, message, code string) {
	writeJSON(w, status, ErrorResponse{Error: message, Code: code})
}
