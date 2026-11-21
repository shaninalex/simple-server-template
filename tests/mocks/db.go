package mocks

import (
	"encoding/json"
	"os"
	"time"
	"uuid"
)

type Post struct {
	ID        uuid.UUID `json:"id"`
	AuthorID  uuid.UUID `json:"author_id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Thumbnail string    `json:"thumbnail"`
	Published time.Time `json:"published"`
	CreatedAt time.Time `json:"created_at"`
}

type Author struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type DB struct {
	Authors []Author `json:"authors"`
	Posts   []Post   `json:"posts"`
}

func newDB() *DB {
	b, err := os.ReadFile("./tests/mocks/data.json")
	if err != nil {
		panic(err)
	}
	var db DB
	if err = json.Unmarshal(b, &db); err != nil {
		panic(err)
	}
	return &db
}
