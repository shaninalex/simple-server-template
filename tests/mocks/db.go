package mocks

import (
	_ "embed"
	"encoding/json"
	"time"
	"uuid"
)

type Post struct {
	ID        uuid.UUID `json:"id"`
	AuthorID  uuid.UUID `json:"author_id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Category  string    `json:"category"`
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

//go:embed data.json
var data []byte

func newDB() *DB {
	var db DB
	if err := json.Unmarshal(data, &db); err != nil {
		panic(err)
	}
	return &db
}
