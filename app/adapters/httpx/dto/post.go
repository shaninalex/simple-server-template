package dto

import (
	"errors"
	"simple-server-template/app/modules/blog/contract"
	"time"
	"uuid"
)

type PostDTO struct {
	ID        uuid.UUID `json:"id"`
	AuthorID  uuid.UUID `json:"author_id"`
	Title     string    `json:"title"`
	Category  string    `json:"category"`
	Content   string    `json:"content"`
	Thumbnail string    `json:"thumbnail"`
	Published time.Time `json:"published"`
	CreatedAt time.Time `json:"created_at"`
}

func (p PostDTO) Verify() error {
	if p.Title == "" {
		return errors.New("post should have a title")
	}
	if len(p.Title) > 200 {
		return errors.New("title is too long")
	}
	if len(p.Content) > 5000 {
		return errors.New("content is too long")
	}
	if p.Published.Before(p.CreatedAt) {
		return errors.New("invalid dates: published before created  at")
	}
	return nil
}

func ToPostDTO(p contract.PostView) PostDTO {
	return PostDTO{
		ID:        p.ID,
		AuthorID:  p.AuthorID,
		Title:     p.Title,
		Category:  p.Category,
		Content:   p.Content,
		Thumbnail: p.Thumbnail,
		Published: p.Published,
		CreatedAt: p.CreatedAt,
	}
}

func ToPostDTOList(posts []contract.PostView) []PostDTO {
	dtos := make([]PostDTO, len(posts))
	for i, p := range posts {
		dtos[i] = ToPostDTO(p)
	}
	return dtos
}

func ToPostView(p PostDTO) contract.PostView {
	return contract.PostView{
		ID:        p.ID,
		AuthorID:  p.AuthorID,
		Title:     p.Title,
		Category:  p.Category,
		Content:   p.Content,
		Thumbnail: p.Thumbnail,
		Published: p.Published,
		CreatedAt: p.CreatedAt,
	}
}
