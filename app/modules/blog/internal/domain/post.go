package domain

import (
	"time"
	"uuid"
)

type Post struct {
	ID        uuid.UUID
	Title     string
	Content   string
	Thumbnail string
	Published time.Time
}
