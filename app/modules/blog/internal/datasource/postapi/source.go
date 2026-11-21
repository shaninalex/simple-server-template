package postapi

import (
	"context"
	"net/url"
	"strconv"

	"simple-server-template/app/modules/blog/internal/domain"
	"simple-server-template/app/platform/httpclient"
)

type remotePost struct {
	ID     int64  `json:"id"`
	UserID int64  `json:"userId"`
	Title  string `json:"title"`
	Body   string `json:"body"`
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

func toDomain(r remotePost) domain.Post {
	return domain.Post{
		ID:      strconv.FormatInt(r.ID, 10),
		Title:   r.Title,
		Content: r.Body,
	}
}
