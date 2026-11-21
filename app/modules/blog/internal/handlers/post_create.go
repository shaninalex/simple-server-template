package handlers

import (
	"context"
	"simple-server-template/app/modules/blog/contract"
	"simple-server-template/app/modules/blog/internal/domain"
	"simple-server-template/app/platform/clock"
)

type CreatePost struct {
	source domain.PostSource
	clk    clock.Clock
}

func NewCreatePost(source domain.PostSource, clk clock.Clock) *CreatePost {
	return &CreatePost{source: source, clk: clk}
}

func (s *CreatePost) Handle(ctx context.Context, cmd contract.CreatePost) (contract.PostView, error) {
	post := toPostDomain(cmd.Post)
	post.CreatedAt = s.clk.Now()
	result, err := s.source.Create(ctx, post)
	if err != nil {
		return contract.PostView{}, err
	}
	return toPostView(result), nil
}
