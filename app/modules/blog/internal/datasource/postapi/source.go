package postapi

import (
	"context"
	"net/url"
	"strconv"
	"time"
	"uuid"

	"simple-server-template/app/modules/blog/internal/domain"
	"simple-server-template/app/platform/httpclient"
)

type remotePost struct {
	ID        uuid.UUID `json:"id"`
	AuthorID  uuid.UUID `json:"author_id"`
	Title     string    `json:"title"`
	Category  string    `json:"category"`
	Content   string    `json:"content"`
	Thumbnail string    `json:"thumbnail"`
	CreatedAt time.Time `json:"created_at"`
	Published time.Time `json:"published"`
}

type Source struct {
	client *httpclient.Client
}

func New(client *httpclient.Client) *Source {
	return &Source{client: client}
}

func (s *Source) List(ctx context.Context, limit, offset int) ([]domain.Post, error) {
	q := url.Values{}
	q.Set("_start", strconv.Itoa(offset))
	q.Set("_limit", strconv.Itoa(limit))

	var remote []remotePost
	if err := s.client.GetJSON(ctx, "/posts", q, &remote); err != nil {
		return nil, err
	}

	posts := make([]domain.Post, len(remote))
	for i, r := range remote {
		posts[i] = toDomain(r)
	}
	return posts, nil
}

func (s *Source) Create(ctx context.Context, post domain.Post) (domain.Post, error) {
	data := remotePost{
		ID:        post.ID,
		AuthorID:  post.AuthorID,
		Title:     post.Title,
		Category:  post.Category,
		Content:   post.Content,
		Thumbnail: post.Thumbnail,
		CreatedAt: post.CreatedAt,
		Published: post.Published,
	}

	var remote remotePost
	if err := s.client.Post(ctx, "/posts", data, &remote); err != nil {
		return domain.Post{}, err
	}
	return toDomain(remote), nil
}

func toDomain(r remotePost) domain.Post {
	return domain.Post{
		ID:        r.ID,
		AuthorID:  r.AuthorID,
		Title:     r.Title,
		Category:  r.Category,
		Content:   r.Content,
		Thumbnail: r.Thumbnail,
		CreatedAt: r.CreatedAt,
		Published: r.Published,
	}
}
