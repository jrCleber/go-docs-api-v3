package routers

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strings"

	handler "codechat.dev/api/handlers"
	"codechat.dev/guards"
	"codechat.dev/internal/domain/instance"
	"codechat.dev/pkg/config"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/sirupsen/logrus"
)

type MsManager struct {
	middle             []any
	instanceMiddleware []handler.InstanceMiddleware
	cfg                *config.Route
	instance           *instance.Service
	logger             *logrus.Entry
}

func NewMsManagerRouter(cfg *config.Route, instance *instance.Service) *MsManager {
	logger := logrus.New()
	return &MsManager{
		cfg:      cfg,
		instance: instance,
		logger:   logger.WithFields(logrus.Fields{"router": "ms-router"}),
	}
}

func (x *MsManager) Auth(middle []any) *MsManager {
	x.middle = middle
	return x
}

func (x *MsManager) InstanceMiddleware(m []handler.InstanceMiddleware) *MsManager {
	x.instanceMiddleware = m
	return x
}

func replacePath(path string) string {
	return strings.Replace(path, "/api/v3", "", 1)
}

func (x *MsManager) redirectPostGet(res *http.Response, err error) *handler.Response {
	response := handler.NewResponse(http.StatusBadRequest)

	if err != nil {
		response.Message = []string{err.Error()}
		return response
	}

	var body any
	err = render.DecodeJSON(res.Body, &body)
	if err != nil {
		response.Message = []string{err.Error()}
		return response
	}

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		if data, ok := body.(map[string]any); ok {
			if d, ok := data["message"].([]any); ok {
				message := d[0].(string)
				err = errors.New(message)
				response.Message = []string{err.Error()}
			} else {
				response.Message = data["message"]
			}

			return response
		}
	}

	response.StatusCode = res.StatusCode

	response.Message = body

	return response
}

func (x *MsManager) redirect(req *http.Request, method string) *handler.Response {
	response := handler.NewResponse(http.StatusBadRequest)

	var bodyBytes []byte
	var err error

	if req.Body != nil {
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			x.logger.Error(err)
		}
		req.Body.Close()
	}

	newBody := io.NopCloser(bytes.NewBuffer(bodyBytes))
	newReq, err := http.NewRequest(method, x.cfg.MsManager+replacePath(req.URL.Path), newBody)
	if err != nil {
		response.Message = []string{err.Error()}
		return response
	}

	newReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(newReq)

	response.StatusCode = res.StatusCode

	if err != nil {
		response.Message = []string{err.Error()}
		return response
	}
	defer res.Body.Close()

	var body any

	err = render.DecodeJSON(res.Body, &body)
	if err != nil {
		response.Message = []string{err.Error()}
		return response
	}

	response.Message = body
	return response
}

func (x *MsManager) Router() *chi.Mux {
	router := chi.NewRouter()

	router.Route("/webhook", func(r chi.Router) {
		r.Use(guards.ApplyGuards(x.middle)...)

		x.logger.Info(x.instanceMiddleware)

		mimetype := "application/json"

		h := handler.NewHandlerMiddleware()

		r.Post("/", h.Middle(x.instanceMiddleware...).ResponseHandler(func(r *http.Request) *handler.Response {
			return x.redirectPostGet(http.Post(x.cfg.MsManager+replacePath(r.URL.Path), mimetype, r.Body))
		}))
		r.Get("/", h.Middle(x.instanceMiddleware...).ResponseHandler(func(r *http.Request) *handler.Response {
			return x.redirectPostGet(http.Get(x.cfg.MsManager + replacePath(r.URL.Path)))
		}))
		r.Get("/{webhookId}", h.Middle(x.instanceMiddleware...).ResponseHandler(func(r *http.Request) *handler.Response {
			return x.redirectPostGet(http.Get(x.cfg.MsManager + replacePath(r.URL.Path)))
		}))
		r.Put("/{webhookId}", h.Middle(x.instanceMiddleware...).ResponseHandler(func(r *http.Request) *handler.Response {
			return x.redirect(r, http.MethodPut)
		}))
		r.Patch("/{webhookId}", h.Middle(x.instanceMiddleware...).ResponseHandler(func(r *http.Request) *handler.Response {
			return x.redirect(r, http.MethodPatch)
		}))
	})

	// router.Route("/qrcode", func(r chi.Router) {
	// 	// r.Use(guards.ApplyGuards(x.middle)...)

	// 	r.Get("/", handler.ResponseHandler(func(r *http.Request) *handler.Response {
	// 		return x.redirectPostGet(http.Get(x.Cfg.MsManager + replacePath(r.URL.Path)))
	// 	}))
	// })

	return router
}
