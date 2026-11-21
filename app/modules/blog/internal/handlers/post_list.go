package handlers

import (
	"context"

	"simple-server-template/app/core/errs"
	"simple-server-template/app/modules/blog/contract"
	"simple-server-template/app/modules/blog/internal/domain"
)

const maxPostListLimit = 100

type PostList struct {
	source domain.PostSource
}

func NewPostList(source domain.PostSource) *PostList {
	return &PostList{source: source}
}

func (s *PostList) Handle(ctx context.Context, query contract.PostList) ([]contract.PostView, error) {
	if query.Limit <= 0 || query.Limit > maxPostListLimit {
		return nil, errs.Validation("BLG01", "limit must be between 1 and 100")
	}
	if query.Offset < 0 {
		return nil, errs.Validation("BLG02", "offset must not be negative")
	}

	posts, err := s.source.List(ctx, query.Limit, query.Offset)
	if err != nil {
		return nil, err
	}

	views := make([]contract.PostView, len(posts))
	for i, p := range posts {
		views[i] = toPostView(p)
	}
	return views, nil
}
