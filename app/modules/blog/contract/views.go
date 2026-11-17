package contract

import (
	"time"
	"uuid"
)

type PostView struct {
	ID        uuid.UUID
	Title     string
	Content   string
	Thumbnail string
	Published time.Time
}
