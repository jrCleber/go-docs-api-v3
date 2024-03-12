package handler

import (
	"net/http"
	"strings"

	"github.com/go-chi/render"
)

type Response struct {
	StatusCode int         `json:"statusCode"`
	Message    interface{} `json:"message,omitempty"`
	Error      interface{} `json:"error"`
}

func NewResponse(status int) *Response {
	return &Response{StatusCode: status}
}

func (r *Response) BadRequest() *Response {
	r.Error = "Bad Request"
	return r
}

func (r *Response) Conflict() *Response {
	r.Error = "Conflict"
	return r
}

func (r *Response) Forbidden() *Response {
	r.Error = "Forbidden"
	return r
}

func (r *Response) InternalServerError() *Response {
	r.Error = "Internal Server Error"
	return r
}

func (r *Response) NotFound() *Response {
	r.Error = "Not Found"
	return r
}

func (r *Response) Unauthorized() *Response {
	r.Error = "Unauthorized"
	return r
}

func (r *Response) Unknown() *Response {
	r.Error = "Unknown Error"
	r.StatusCode = 520
	return r
}

func (r *Response) ObjectEmpty() *Response {
	r.StatusCode = http.StatusBadRequest
	r.Error = []any{"The object cannot be empty."}
	return r
}

func (r *Response) GetResponse() any {
	switch r.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusNoContent:
		return r.GetData()
	case http.StatusBadRequest:
		return r.BadRequest()
	case http.StatusConflict:
		return r.Conflict()
	case http.StatusForbidden:
		return r.Forbidden()
	case http.StatusInternalServerError:
		return r.InternalServerError()
	case http.StatusNotFound:
		return r.NotFound()
	case http.StatusUnauthorized:
		return r.Unauthorized()
	default:
		return r.Unknown()
	}
}

func (r *Response) GetData() any {
	return r.Message
}

func extractErrorDetails(errMsg string) (fieldName string, dataType string) {
	fieldIndex := strings.Index(errMsg, "field ")
	ofIndex := strings.Index(errMsg, " of")
	if fieldIndex == -1 || ofIndex == -1 {
		return "", ""
	}
	fieldName = errMsg[fieldIndex+6 : ofIndex]

	typeIndex := strings.Index(errMsg, "type ")
	if typeIndex == -1 {
		return fieldName, ""
	}
	dataType = errMsg[typeIndex+5:]

	return fieldName, dataType
}

func UnmarshalDescriptionError(e error) *Response {
	if e != nil {
		fieldName, dataType := extractErrorDetails(e.Error())
		return &Response{
			StatusCode: http.StatusBadRequest,
			Message:    []string{fieldName + " must be of type " + dataType + "."},
		}
	}

	return nil
}

type HandlerFunc func(r *http.Request) *Response
type InstanceMiddleware func(w http.ResponseWriter, r *http.Request) bool

type HandlerMiddleware struct {
	middle []InstanceMiddleware
}

func (h *HandlerMiddleware) Middle(m ...InstanceMiddleware) *HandlerMiddleware {
	h.middle = m
	return h
}

func (h *HandlerMiddleware) ResponseHandler(handle HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		for _, f := range h.middle {
			next := f(w, r)
			if !next {
				return
			}
		}

		handler := handle(r)
		render.Status(r, handler.StatusCode)
		render.JSON(w, r, handler.GetResponse())
	}
}

func NewHandlerMiddleware() *HandlerMiddleware {
	return &HandlerMiddleware{}
}

func ErrorList(errs []error) *[]any {
	list := make([]any, len(errs))
	for i, v := range errs {
		list[i] = v.Error()
	}

	return &list
}
