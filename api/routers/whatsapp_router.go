package routers

import (
	handler "codechat.dev/api/handlers"
	"codechat.dev/guards"
	"github.com/go-chi/chi/v5"
)

type Whatsapp struct {
	handler *handler.Instance
	middle  []any
	instanceMiddleware []handler.InstanceMiddleware
}

func NewWhatsAppRouter(handler *handler.Instance) *Whatsapp {
	return &Whatsapp{handler: handler}
}

func (x *Whatsapp) Auth(middle []any) *Whatsapp {
	x.middle = middle
	return x
}

func (x *Whatsapp) InstanceMiddleware(m []handler.InstanceMiddleware) *Whatsapp {
	x.instanceMiddleware = m
	return x
}

func (x *Whatsapp) Router() *chi.Mux {
	router := chi.NewRouter()

	router.Route("/", func(r chi.Router) {
		r.Use(guards.ApplyGuards(x.middle)...)

		h := handler.NewHandlerMiddleware()

		r.Get("/connect",h.Middle(x.instanceMiddleware...).ResponseHandler(x.handler.GetInstanceConnect))
		r.Patch("/name",h.Middle(x.instanceMiddleware...).ResponseHandler(x.handler.PatchUpdateProfileName))
		r.Patch("/logout",h.Middle(x.instanceMiddleware...).ResponseHandler(x.handler.PatchInstanceLogout))
	})

	return router
}
