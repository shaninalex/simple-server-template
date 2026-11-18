package api

import (
	"simple-server-template/app/adapters/api/controllers"
	"simple-server-template/app/adapters/httpx"
	"simple-server-template/app/core"
)

func RegisterAPIControllers(router *httpx.Mux, resolve core.Resolve) {
	controllers.RegisterPostController(router, resolve)
}
