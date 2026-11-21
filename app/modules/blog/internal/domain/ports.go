package domain

import (
	"context"
	"uuid"
)

type PostSource interface {
	List(ctx context.Context, limit, offset int) ([]Post, error)
	Create(ctx context.Context, post Post) (Post, error)
	Get(ctx context.Context, id uuid.UUID) (Post, error)
}
