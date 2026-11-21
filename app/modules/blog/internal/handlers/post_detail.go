package handlers

import (
	"context"
	"simple-server-template/app/modules/blog/contract"
	"simple-server-template/app/modules/blog/internal/domain"
)

type PostDetail struct {
	source domain.PostSource
}

func NewPostDetail(source domain.PostSource) *PostDetail {
	return &PostDetail{source}
}

func (s *PostDetail) Handle(ctx context.Context, q contract.PostDetail) (contract.PostView, error) {
	post, err := s.source.Get(ctx, q.ID)
	if err != nil {
		return contract.PostView{}, err
	}
	return toPostView(post), nil
}
