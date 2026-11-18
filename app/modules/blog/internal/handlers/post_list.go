package handlers

import (
	"context"
	"simple-server-template/app/modules/blog/contract"
	"time"
	"uuid"
)

type PostList struct {
}

func NewPostList() *PostList {
	return &PostList{}
}

func (s *PostList) Handle(ctx context.Context, query contract.PostList) ([]contract.PostView, error) {
	// TODO: call datasource here via repository or data provider
	// Handlers is where real work happening.
	var posts = []contract.PostView{
		{
			ID:        uuid.New(),
			Title:     "The Future of Web Development",
			Content:   "Modern web development is evolving rapidly, with new tools and frameworks changing how developers build applications.",
			Thumbnail: "https://picsum.photos/seed/webdev/800/600",
			Published: time.Now().AddDate(0, 0, -1),
		},
		{
			ID:        uuid.New(),
			Title:     "Understanding Clean Architecture",
			Content:   "Clean architecture helps developers build maintainable, scalable, and testable software by separating business logic from infrastructure.",
			Thumbnail: "https://picsum.photos/seed/architecture/800/600",
			Published: time.Now().AddDate(0, 0, -3),
		},
		{
			ID:        uuid.New(),
			Title:     "Why Go Is Worth Learning",
			Content:   "Go combines simplicity, performance, and excellent concurrency support, making it a great choice for backend development.",
			Thumbnail: "https://picsum.photos/seed/golang/800/600",
			Published: time.Now().AddDate(0, 0, -5),
		},
		{
			ID:        uuid.New(),
			Title:     "Building Better User Experiences",
			Content:   "A great user experience is built on simplicity, accessibility, consistency, and understanding what users actually need.",
			Thumbnail: "https://picsum.photos/seed/uxdesign/800/600",
			Published: time.Now().AddDate(0, 0, -8),
		},
		{
			ID:        uuid.New(),
			Title:     "The Power of Open Source",
			Content:   "Open source software brings developers together to collaborate, share knowledge, and build tools that benefit everyone.",
			Thumbnail: "https://picsum.photos/seed/opensource/800/600",
			Published: time.Now().AddDate(0, -1, 0),
		},
		{
			ID:        uuid.New(),
			Title:     "Getting Started with Docker",
			Content:   "Docker makes application deployment predictable by packaging applications and their dependencies into portable containers.",
			Thumbnail: "https://picsum.photos/seed/docker/800/600",
			Published: time.Now().AddDate(0, -1, -5),
		},
		{
			ID:        uuid.New(),
			Title:     "A Guide to Database Optimization",
			Content:   "Efficient database queries, proper indexing, and thoughtful schema design can significantly improve application performance.",
			Thumbnail: "https://picsum.photos/seed/database/800/600",
			Published: time.Now().AddDate(0, -2, 0),
		},
		{
			ID:        uuid.New(),
			Title:     "The Art of Minimalist Design",
			Content:   "Minimalist design focuses on clarity and purpose, removing unnecessary elements while keeping the experience intuitive.",
			Thumbnail: "https://picsum.photos/seed/minimalism/800/600",
			Published: time.Now().AddDate(0, -3, 0),
		},
	}
	return posts, nil
}
