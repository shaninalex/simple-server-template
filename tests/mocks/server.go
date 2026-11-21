package mocks

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"uuid"
)

func New() *httptest.Server {
	return httptest.NewServer(Handler())
}

func Handler() http.Handler {
	mux := http.NewServeMux()
	s := mockHandlers{DB: newDB()}

	mux.HandleFunc("GET /posts", s.listPosts)
	mux.HandleFunc("GET /posts/{id}", s.getPost)
	return mux
}

type mockHandlers struct {
	DB *DB
}

func (s *mockHandlers) listPosts(w http.ResponseWriter, r *http.Request) {
	start := queryInt(r, "_start", 0)
	limit := queryInt(r, "_limit", len(s.DB.Posts))

	start = min(max(start, 0), len(s.DB.Posts))
	end := min(start+max(limit, 0), len(s.DB.Posts))

	writeJSON(w, http.StatusOK, s.DB.Posts[start:end])
}

func (s *mockHandlers) getPost(w http.ResponseWriter, r *http.Request) {
	id := uuid.MustParse(r.PathValue("id"))
	for _, p := range s.DB.Posts {
		if p.ID == id {
			writeJSON(w, http.StatusOK, p)
			return
		}
	}
	writeJSON(w, http.StatusNotFound, struct{}{})
}

func queryInt(r *http.Request, key string, fallback int) int {
	v, err := strconv.Atoi(r.URL.Query().Get(key))
	if err != nil {
		return fallback
	}
	return v
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
