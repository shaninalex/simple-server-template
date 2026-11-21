package handlers

import (
	"simple-server-template/app/modules/blog/contract"
	"simple-server-template/app/modules/blog/internal/domain"
)

func toPostView(p domain.Post) contract.PostView {
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
func toPostDomain(p contract.PostView) domain.Post {
	return domain.Post{
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
