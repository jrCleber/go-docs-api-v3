package handler

import (
	"net/http"

	"codechat.dev/contract"
	"codechat.dev/internal/domain/chat"
	"codechat.dev/pkg/validate"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/sirupsen/logrus"
)

type Chat struct {
	service *chat.Service
	logger  logrus.Entry
}

func NewChat(service *chat.Service) *Chat {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	return &Chat{
		service: service,
		logger:  *logger.WithFields(logrus.Fields{"handler": "chat"}),
	}
}

func (h *Chat) CheckNumbers(r *http.Request) *Response {
	var body contract.ChatNumbers
	e := render.DecodeJSON(r.Body, &body)
	if err := UnmarshalDescriptionError(e); err != nil {
		return err
	}

	response := NewResponse(http.StatusBadRequest)

	errs := validate.Struct(body)
	if errs != nil {
		response.Message = ErrorList(errs)
		return response
	}

	param := chi.URLParam(r, "instance")
	data, status, err := h.service.ValidateWhatsAppNumbers(param, body.Numbers)

	response.StatusCode = status

	if err != nil {
		response.Message = []any{err.Error()}
	}

	response.Message = data
	return response
}
