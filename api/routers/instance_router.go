package routers

import (
	handler "codechat.dev/api/handlers"
	"codechat.dev/guards"
	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
)

type Instance struct {
	handler            *handler.Instance
	authMiddle         []any
	instanceMiddleware []handler.InstanceMiddleware
	rootParam          string
	rootPath           string
	logger             *logrus.Entry
}

func NewInstanceRouter(handler *handler.Instance) *Instance {
	logger := logrus.New()
	return &Instance{
		handler: handler,
		logger:  logger.WithFields(logrus.Fields{"name": "instance-route"}),
	}
}

func (x *Instance) Auth(middle ...any) *Instance {
	x.authMiddle = middle
	return x
}

func (x *Instance) GlobalMiddleware(m ...handler.InstanceMiddleware) *Instance {
	x.instanceMiddleware = m
	return x
}

func (x *Instance) RootParam(param string) *Instance {
	x.rootParam = param
	return x
}

func (x *Instance) RootPath(path string) *Instance {
	x.rootPath = path
	return x
}

func (x *Instance) Routers(routes ...InnerRouters) *chi.Mux {
	router := chi.NewRouter()

	router.Route(x.rootPath, func(r chi.Router) {
		r.Use(guards.ApplyGuards(x.authMiddle)...)

		h := handler.NewHandlerMiddleware()

		r.Post("/", h.Middle(x.instanceMiddleware...).ResponseHandler(x.handler.PostInstanceCreate))
		r.Get("/", h.Middle(x.instanceMiddleware...).ResponseHandler(x.handler.GetAllInstance))
		r.Get(x.rootParam, h.Middle(x.instanceMiddleware...).ResponseHandler(x.handler.GetInstance))
		r.Delete(x.rootParam, h.Middle(x.instanceMiddleware...).ResponseHandler(x.handler.DeleteInstance))
	})

	for _, v := range routes {
		r := v.Router
		m := v.InstanceMiddleware
		path := x.rootPath + x.rootParam + v.Path
		switch v := r.(type) {
		case *SendMessage:
			v.Auth(x.authMiddle)
			m = append(m, v.instanceMiddleware...)
			v.InstanceMiddleware(m)
			router.Mount(path, v.Router())
		case *Whatsapp:
			v.Auth(x.authMiddle)
			m = append(m, v.instanceMiddleware...)
			v.InstanceMiddleware(m)
			router.Mount(path, v.Router())
		case *MsManager:
			v.Auth(x.authMiddle)
			m = append(m, v.instanceMiddleware...)
			v.InstanceMiddleware(m)
			router.Mount(path, v.Router())
		}
	}

	return router
}
