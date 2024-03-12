package handler

import (
	"net/http"

	"codechat.dev/contract"
	"codechat.dev/internal/domain/instance"
	"codechat.dev/pkg/validate"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/sirupsen/logrus"
)

type Instance struct {
	service *instance.Service
	logger  logrus.Entry
}

func NewInstance(service *instance.Service) *Instance {
	logger := logrus.New()
	return &Instance{
		service: service,
		logger:  *logger.WithFields(logrus.Fields{"handler": "instance"}),
	}
}

func (h *Instance) PostInstanceCreate(r *http.Request) *Response {
	var body contract.Instance
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

	data, status, err := h.service.Create(&body)
	response.StatusCode = status
	if err != nil {
		response.Message = []any{
			"Unable to create instance",
			err.Error(),
		}
		response.StatusCode = http.StatusBadRequest
		return response
	}

	response.Message = data
	return response
}

func (h *Instance) GetInstance(r *http.Request) *Response {
	param := chi.URLParam(r, "instance")

	h.logger.Info(param)

	data, status, err := h.service.Find(param)

	response := NewResponse(status)

	if err != nil {
		response.Message = []any{err.Error()}
		return response
	}

	response.Message = data
	return response
}

func (h *Instance) GetAllInstance(r *http.Request) *Response {
	data, status, err := h.service.FindAll()

	response := NewResponse(status)

	if err != nil {
		response.Message = []string{err.Error()}
		return response
	}

	response.Message = data
	return response
}

func (h *Instance) GetInstanceConnect(r *http.Request) *Response {
	param := chi.URLParam(r, "instance")

	data, status, _, err := h.service.NewConnection(param)

	response := NewResponse(status)

	if err != nil {
		response.Message = []any{
			"The connection could not be completed.",
			err.Error(),
		}
		return response
	}

	response.Message = data
	return response
}

func (h *Instance) PatchInstanceLogout(r *http.Request) *Response {
	param := chi.URLParam(r, "instance")

	data, status, err := h.service.Logout(param)

	response := NewResponse(status)

	if err != nil {
		response.Error = []any{
			"Unable to log out the instance.",
			err.Error(),
		}
		return response
	}

	response.Message = data
	return response
}

func (h *Instance) DeleteInstance(r *http.Request) *Response {
	param := chi.URLParam(r, "instance")

	data, status, err := h.service.Delete(param)

	response := NewResponse(status)

	if err != nil {
		response.Message = []any{err.Error()}
		return response
	}

	response.Message = data
	return response
}

func (h *Instance) PatchUpdateProfileName(r *http.Request) *Response {
	var body contract.InstanceName

	e := render.DecodeJSON(r.Body, &body)
	if err := UnmarshalDescriptionError(e); err != nil {
		return err
	}

	response := NewResponse(200)

	errs := validate.Struct(&body)
	if errs != nil {
		list := make([]any, len(errs))
		for i, v := range errs {
			list[i] = v.Error()
		}
		response.Message = list
		return response.BadRequest()
	}

	param := chi.URLParam(r, "instance")

	status, err := h.service.UpdateProfileName(param, body.Name)
	response.StatusCode = status

	if err != nil {
		response.Message = []any{
			"Unable to update name",
			err.Error(),
		}
		return response
	}

	return response
}
