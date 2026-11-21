package postsapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
)

type Post struct {
	UserID int64  `json:"userId"`
	ID     int64  `json:"id"`
	Title  string `json:"title"`
	Body   string `json:"body"`
}

var Posts = []Post{
	{UserID: 1, ID: 1, Title: "The Future of Web Development", Body: "Modern web development is evolving rapidly, with new tools and frameworks changing how developers build applications."},
	{UserID: 1, ID: 2, Title: "Understanding Clean Architecture", Body: "Clean architecture helps developers build maintainable, scalable, and testable software by separating business logic from infrastructure."},
	{UserID: 1, ID: 3, Title: "Why Go Is Worth Learning", Body: "Go combines simplicity, performance, and excellent concurrency support, making it a great choice for backend development."},
	{UserID: 2, ID: 4, Title: "Building Better User Experiences", Body: "A great user experience is built on simplicity, accessibility, consistency, and understanding what users actually need."},
	{UserID: 2, ID: 5, Title: "The Power of Open Source", Body: "Open source software brings developers together to collaborate, share knowledge, and build tools that benefit everyone."},
	{UserID: 2, ID: 6, Title: "Getting Started with Docker", Body: "Docker makes application deployment predictable by packaging applications and their dependencies into portable containers."},
	{UserID: 3, ID: 7, Title: "A Guide to Database Optimization", Body: "Efficient database queries, proper indexing, and thoughtful schema design can significantly improve application performance."},
	{UserID: 3, ID: 8, Title: "The Art of Minimalist Design", Body: "Minimalist design focuses on clarity and purpose, removing unnecessary elements while keeping the experience intuitive."},
}

func New() *httptest.Server {
	return httptest.NewServer(Handler())
}

func Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /posts", listPosts)
	mux.HandleFunc("GET /posts/{id}", getPost)
	return mux
}

func listPosts(w http.ResponseWriter, r *http.Request) {
	start := queryInt(r, "_start", 0)
	limit := queryInt(r, "_limit", len(Posts))

	start = min(max(start, 0), len(Posts))
	end := min(start+max(limit, 0), len(Posts))

	writeJSON(w, http.StatusOK, Posts[start:end])
}

func getPost(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err == nil {
		for _, p := range Posts {
			if p.ID == id {
				writeJSON(w, http.StatusOK, p)
				return
			}
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
