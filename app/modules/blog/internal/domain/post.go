package domain

import (
	"time"
)

type Post struct {
	ID        string
	Title     string
	Content   string
	Thumbnail string
	Published time.Time
}
