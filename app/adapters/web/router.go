package web

import (
	"net/http"
	"simple-server-template/app/adapters/httpx"
	"simple-server-template/app/adapters/httpx/dto"
	"simple-server-template/app/adapters/web/templates"
	"simple-server-template/app/core"
	"simple-server-template/app/modules/blog/contract"
)

func RegisterWebRoutes(router *httpx.Mux, resolve core.Resolve) {
	router.Handle("GET /", handleHomePage(resolve))
}

func handleHomePage(resolve core.Resolve) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		posts, err := contract.AskPostList(r.Context(), resolve(), contract.PostList{
			Limit:  10,
			Offset: 0,
		})
		if err != nil {
			httpx.ErrorResponse(w, httpx.StatusFromError(err), err)
			return
		}

		component := templates.HomePage(dto.ToPostDTOList(posts))
		component.Render(r.Context(), w)
	})
}
