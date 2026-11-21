package contract

import (
	"time"
	"uuid"
)

type PostView struct {
	ID        uuid.UUID
	AuthorID  uuid.UUID
	Title     string
	Category  string
	Content   string
	Thumbnail string
	Published time.Time
	CreatedAt time.Time
}
