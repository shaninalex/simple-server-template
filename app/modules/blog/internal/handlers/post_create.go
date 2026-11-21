package handlers

import (
	"context"
	"simple-server-template/app/modules/blog/contract"
	"uuid"
)

type CreatePost struct {
	source PostSource
}

func NewCreatePost(source PostSource) *CreatePost {
	return &CreatePost{source: source}

}

func (s *CreatePost) Handle(ctx context.Context, cmd contract.CreatePost) (contract.PostView, error) {
	cmd.Post.ID = uuid.New()
	return cmd.Post, nil
}
