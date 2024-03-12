package routers

import (
	"net/http"
	"strings"

	handler "codechat.dev/api/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type InnerRouters struct {
	Path               string
	Router             any
	InstanceMiddleware []handler.InstanceMiddleware
}

type RouterInterface interface {
	Middlewares(middlewares []any) any
	Router() *chi.Mux
}

func PingOptions(r *chi.Mux) {
	r.Options("/ping", func(w http.ResponseWriter, r *http.Request) {
		response := handler.NewResponse(http.StatusOK)
		response.Message = "pong"

		render.Status(r, response.StatusCode)
		render.JSON(w, r, response.GetData())
	})
}

func NotFound(r *chi.Mux) {
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		method := strings.ToUpper(r.Method)
		path := r.URL.Path

		response := handler.NewResponse(http.StatusNotFound)
		response.Message = "Cannot " + method + " " + path

		render.Status(r, response.StatusCode)
		render.JSON(w, r, response.GetResponse())
	})
}
