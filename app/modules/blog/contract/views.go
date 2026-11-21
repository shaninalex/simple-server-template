package contract

import (
	"time"
)

type PostView struct {
	ID        string
	Title     string
	Content   string
	Thumbnail string
	Published time.Time
}
