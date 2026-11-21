package httpx

import (
	"encoding/json"
	"net/http"
)

type Validatable interface {
	Verify() error
}

func BindPayload[T Validatable](req *http.Request, payload T) (*T, error) {
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		return nil, err
	}
	if err := payload.Verify(); err != nil {
		return nil, err
	}

	return &payload, nil
}
