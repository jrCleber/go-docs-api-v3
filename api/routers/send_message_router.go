package routers

import (
	handler "codechat.dev/api/handlers"
	"codechat.dev/guards"
	"github.com/go-chi/chi/v5"
)

type SendMessage struct {
	handler            *handler.SendMessage
	middle             []any
	instanceMiddleware []handler.InstanceMiddleware
}

func NewSendMessageRouter(handler *handler.SendMessage) *SendMessage {
	return &SendMessage{handler: handler}
}

func (x *SendMessage) Auth(middle []any) *SendMessage {
	x.middle = middle
	return x
}

func (x *SendMessage) InstanceMiddleware(m []handler.InstanceMiddleware) *SendMessage {
	x.instanceMiddleware = m
	return x
}

func (x *SendMessage) Router() *chi.Mux {
	router := chi.NewRouter()

	router.Route("/", func(r chi.Router) {
		r.Use(guards.ApplyGuards(x.middle)...)

		h := handler.NewHandlerMiddleware()

		r.Post("/text", h.Middle(x.instanceMiddleware...).ResponseHandler(x.handler.PostSendText))
		r.Post("/link-preview", h.Middle(x.instanceMiddleware...).ResponseHandler(x.handler.PostSendLink))
		r.Post("/media", h.Middle(x.instanceMiddleware...).ResponseHandler(x.handler.PostSendMediaUrl))
		r.Post("/media-file", h.Middle(x.instanceMiddleware...).ResponseHandler(x.handler.PostSenMediaFile))
		r.Post("/audio", h.Middle(x.instanceMiddleware...).ResponseHandler(x.handler.PostSendMediaUrl))
		r.Post("/audio-file", h.Middle(x.instanceMiddleware...).ResponseHandler(x.handler.PostSenMediaFile))
		r.Post("/ptv", h.Middle(x.instanceMiddleware...).ResponseHandler(x.handler.PostSendMediaUrl))
		r.Post("/ptv-file", h.Middle(x.instanceMiddleware...).ResponseHandler(x.handler.PostSenMediaFile))
		r.Post("/location", h.Middle(x.instanceMiddleware...).ResponseHandler(x.handler.PostSendLocation))
		r.Post("/contact", h.Middle(x.instanceMiddleware...).ResponseHandler(x.handler.PostSendContact))
		r.Post("/list", h.Middle(x.instanceMiddleware...).ResponseHandler(x.handler.PostSendList))
		r.Post("/poll", h.Middle(x.instanceMiddleware...).ResponseHandler(x.handler.PostSendPoll))
		r.Patch("/reaction", h.Middle(x.instanceMiddleware...).ResponseHandler(x.handler.PatchSendReaction))
		r.Patch("/edit", h.Middle(x.instanceMiddleware...).ResponseHandler(x.handler.PatchEditMessage))
	})

	return router
}
