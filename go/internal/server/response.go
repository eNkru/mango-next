package server

import (
	"encoding/json"
	"net/http"
)

func sendJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func sendJSONError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]any{
		"success": false,
		"error":   msg,
	})
}

func sendText(w http.ResponseWriter, text string) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(text))
}

func sendAttachment(w http.ResponseWriter, data []byte, filename, mime string) {
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	w.Header().Set("Content-Type", mime)
	w.Write(data)
}

func sendImage(w http.ResponseWriter, data []byte, mime string) {
	w.Header().Set("Content-Type", mime)
	w.Write(data)
}

func redirect(w http.ResponseWriter, r *http.Request, path string) {
	http.Redirect(w, r, path, http.StatusFound)
}
