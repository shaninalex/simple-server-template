package controllers

import (
	"net/http"
	"simple-server-template/app/adapters/httpx"
	"simple-server-template/app/adapters/httpx/dto"
	"simple-server-template/app/core"
	"simple-server-template/app/modules/blog/contract"
)

func RegisterPostController(router *httpx.Mux, resolve core.Resolve) {
	router.Handle("GET /api/posts", PostList(resolve))
	router.Handle("POST /api/posts", PostCreate(resolve))
}

func PostList(resolve core.Resolve) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		posts, err := contract.AskPostList(r.Context(), resolve(), contract.PostList{
			Limit:  10,
			Offset: 0,
		})
		if err != nil {
			httpx.ErrorResponse(w, httpx.StatusFromError(err), err)
			return
		}
		httpx.JSONResponse(w, 200, dto.ToPostDTOList(posts))
	})
}

func PostCreate(resolve core.Resolve) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bindPayload, err := httpx.BindPayload(r, dto.PostDTO{})
		if err != nil {
			httpx.ErrorResponse(w, 400, err)
			return
		}
		post, err := contract.ExecCreatePost(r.Context(), resolve(), contract.CreatePost{
			Post: dto.ToPostView(*bindPayload),
		})
		if err != nil {
			httpx.ErrorResponse(w, httpx.StatusFromError(err), err)
			return
		}
		httpx.JSONResponse(w, http.StatusCreated, dto.ToPostDTO(post))
	})
}
