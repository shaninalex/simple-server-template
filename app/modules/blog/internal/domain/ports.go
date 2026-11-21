package domain

import "context"

type PostSource interface {
	List(ctx context.Context, limit, offset int) ([]Post, error)
	Create(ctx context.Context, post Post) (Post, error)
}
