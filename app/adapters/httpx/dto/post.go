package dto

import (
	"simple-server-template/app/modules/blog/contract"
	"time"
	"uuid"
)

type PostDTO struct {
	ID        uuid.UUID `json:"id"`
	AuthorID  uuid.UUID `json:"author_id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Thumbnail string    `json:"thumbnail"`
	Published time.Time `json:"published"`
	CreatedAt time.Time `json:"created_at"`
}

func (p PostDTO) Verify() error {
	return nil
}

func ToPostDTO(p contract.PostView) PostDTO {
	return PostDTO{
		ID:        p.ID,
		AuthorID:  p.AuthorID,
		Title:     p.Title,
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
		Content:   p.Content,
		Thumbnail: p.Thumbnail,
		Published: p.Published,
		CreatedAt: p.CreatedAt,
	}
}
