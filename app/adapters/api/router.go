package api

import (
	"net/http"
	"simple-server-template/app/adapters/api/controllers"
	"simple-server-template/app/core"
)

func RegisterAPIControllers(router *http.ServeMux, resolve core.Resolve) {
	controllers.RegisterPostController(router, resolve)
}
