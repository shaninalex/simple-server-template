package domain

import (
	"time"
	"uuid"
)

type Post struct {
	ID        uuid.UUID
	AuthorID  uuid.UUID
	Category  string
	Title     string
	Content   string
	Thumbnail string
	CreatedAt time.Time
	Published time.Time
}
