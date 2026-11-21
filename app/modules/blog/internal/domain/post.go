package domain

import (
	"time"
	"uuid"
)

type Post struct {
	ID        uuid.UUID
	AuthorID  uuid.UUID
	Title     string
	Content   string
	Thumbnail string
	CreatedAt time.Time
	Published time.Time
}
