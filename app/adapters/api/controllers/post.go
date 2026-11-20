package controllers

import (
	"net/http"
	"simple-server-template/app/adapters/httpx"
	"simple-server-template/app/adapters/httpx/dto"
	"simple-server-template/app/core"
	"simple-server-template/app/modules/blog/contract"
)

func RegisterPostController(router *httpx.Mux, resolve core.Resolve) {
	router.Handle("GET /api/posts", postList(resolve))
}

func postList(resolve core.Resolve) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		posts, err := contract.AskPostList(r.Context(), resolve(), contract.PostList{
			Limit:  10,
			Offset: 0,
		})
		if err != nil {
			httpx.ErrorResponse(w, 400, err)
			return
		}
		httpx.JSONResponse(w, 200, dto.ToPostDTOList(posts))
	})
}
