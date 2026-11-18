package controllers

import (
	"simple-server-template/app/modules/blog/contract"
	"time"
	"uuid"
)

type PostDTO struct {
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Thumbnail string    `json:"thumbnail"`
	Published time.Time `json:"published"`
}

func toPostDTO(p contract.PostView) PostDTO {
	return PostDTO{
		ID:        p.ID,
		Title:     p.Title,
		Content:   p.Content,
		Thumbnail: p.Thumbnail,
		Published: p.Published,
	}
}

func toPostDTOList(posts []contract.PostView) []PostDTO {
	dtos := make([]PostDTO, len(posts))
	for i, p := range posts {
		dtos[i] = toPostDTO(p)
	}
	return dtos
}
