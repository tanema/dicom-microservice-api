package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type errorResponse struct {
	Error string `json:"error"`
}

func jsonErr(w http.ResponseWriter, status int, msg string) {
	data, err := json.Marshal(errorResponse{Error: msg})
	if err != nil {
		http.Error(w, fmt.Sprintf("problem marshalling json error response: %v -> orig error %v", err, msg), http.StatusInternalServerError)
	}
	w.Header().Add("Content-Type", "application/json")
	http.Error(w, string(data), status)
}

func jsonOk(w http.ResponseWriter, data any) {
	body, err := json.Marshal(data)
	if err != nil {
		http.Error(w, fmt.Sprintf("problem marshalling json response: %v", err), http.StatusInternalServerError)
	}
	w.Header().Add("Content-Type", "application/json")
	_, _ = fmt.Fprint(w, string(body))
}
