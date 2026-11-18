package httpx

import (
	"bytes"
	"encoding/json"
	"net/http"
)

func JSONResponse(w http.ResponseWriter, status int, data any) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = buf.WriteTo(w)
}

type ErrorResponseData struct {
	Error string `json:"error"`
}

func NewErrorResponseData(err error) ErrorResponseData {
	return ErrorResponseData{Error: err.Error()}
}

func ErrorResponse(w http.ResponseWriter, status int, err error) {
	JSONResponse(w, status, NewErrorResponseData(err))
}
