package handlers

import (
	"context"
	"simple-server-template/app/modules/blog/contract"
)

type PostList struct {
}

func NewPostList() *PostList {
	return &PostList{}
}

func (s *PostList) Handle(ctx context.Context, query contract.PostList) ([]contract.PostView, error) {
	// TODO: call datasource here via repository or data provider
	// Handlers is where real work happening.
	return []contract.PostView{}, nil
}
