package handler

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"codechat.dev/contract"
	sendmessage "codechat.dev/internal/domain/send_message"
	"codechat.dev/pkg/validate"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/sirupsen/logrus"
)

type SendMessage struct {
	service *sendmessage.Service
	logger  logrus.Entry
}

func validateQuoted(ref interface{}) (*contract.Quoted, []error) {
	var err []error

	if ref != nil {
		q, ok := ref.(*contract.Quoted)
		if !ok {
			err = append(err, errors.New("quote message cannot be validated"))
			return nil, err
		}

		err = validate.Struct(q)
		return q, err
	}

	return nil, nil
}

func NewSendMessage(service *sendmessage.Service) *SendMessage {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	return &SendMessage{
		service: service,
		logger:  *logger.WithFields(logrus.Fields{"handler": "sendMessage"}),
	}
}

func (h *SendMessage) PostSendText(r *http.Request) *Response {
	var body contract.TextMessage
	e := render.DecodeJSON(r.Body, &body)
	if err := UnmarshalDescriptionError(e); err != nil {
		return err
	}

	response := NewResponse(http.StatusBadRequest)

	if errs := validate.Struct(body); errs != nil {
		response.Message = ErrorList(errs)
		return response
	}

	quoted, errs := validateQuoted(body.Options.QuotedMessage)
	if errs != nil {
		response.Message = ErrorList(errs)
		return response
	}

	param := chi.URLParam(r, "instance")
	data, status, err := h.service.TextMessage(param, &body, quoted)

	response.StatusCode = status

	if err != nil {
		response.Message = []any{err.Error()}
		return response
	}

	response.Message = data
	return response
}

func (h *SendMessage) PostSendLink(r *http.Request) *Response {
	var body contract.LinkMessage
	e := render.DecodeJSON(r.Body, &body)
	if err := UnmarshalDescriptionError(e); err != nil {
		return err
	}

	response := NewResponse(http.StatusBadRequest)

	if errs := validate.Struct(body); errs != nil {
		response.Message = ErrorList(errs)
		return response
	}

	quoted, errs := validateQuoted(body.Options.QuotedMessage)
	if errs != nil {
		response.Message = ErrorList(errs)
		return response
	}

	param := chi.URLParam(r, "instance")
	data, status, err := h.service.LinkMessage(param, &body, quoted)

	response.StatusCode = status

	if err != nil {
		response.Message = []any{err.Error()}
		return response
	}

	response.Message = data
	return response
}

func (h *SendMessage) PostSendMediaUrl(r *http.Request) *Response {
	var b contract.MediaMessage

	response := NewResponse(http.StatusBadRequest)

	var mime string
	if strings.Contains(r.URL.Path, "/send/audio") {
		var body contract.AudioMessage
		e := render.DecodeJSON(r.Body, &body)
		if err := UnmarshalDescriptionError(e); err != nil {
			return err
		}

		mime = "audio/ogg; codecs=opus"
		if body.Options.Presence != "recording" {
			body.Options.Presence = "none"
		}

		b.Options = body.Options
		b.Message = body.Message
		b.Message.MediaType = "audio"
		b.Recipient = body.Recipient
	} else if strings.Contains(r.URL.Path, "/send/ptv") {
		var body contract.PtvMessage
		e := render.DecodeJSON(r.Body, &body)
		if err := UnmarshalDescriptionError(e); err != nil {
			return err
		}

		if body.Options.Presence != "recording" {
			body.Options.Presence = "none"
		}

		b.Options = body.Options
		b.Message = body.Message
		b.Message.MediaType = "ptv"
		b.Recipient = body.Recipient
	} else {
		e := render.DecodeJSON(r.Body, &b)
		if err := UnmarshalDescriptionError(e); err != nil {
			return err
		}
	}

	if errs := validate.Struct(b); errs != nil {
		response.Message = ErrorList(errs)
		return response
	}

	quoted, errs := validateQuoted(b.Options.QuotedMessage)
	if errs != nil {
		response.Message = ErrorList(errs)
		return response
	}

	param := chi.URLParam(r, "instance")

	data, status, err := h.service.MediaMessage(param, mime, &b, quoted)
	response.StatusCode = status
	if err != nil {
		response.Message = []any{err.Error()}
		return response
	}

	response.Message = data
	return response
}

func (h *SendMessage) PostSenMediaFile(r *http.Request) *Response {
	response := NewResponse(http.StatusBadRequest)

	err := r.ParseMultipartForm(16 << 20)
	if err != nil {
		response.Message = []any{
			"Failed to parse the form",
			err.Error(),
		}
		return response
	}

	var body contract.MediaMessage

	body.RecipientParam = contract.RecipientParam{Recipient: r.FormValue("recipient")}
	body.Options = contract.Options{
		Presence:           r.FormValue("presence"),
		ExternalAttributes: r.FormValue("externalAttributes"),
	}

	delay, err := strconv.ParseInt(r.FormValue("delay"), 10, 64)
	body.Options.Delay = int(delay)
	if err != nil {
		body.Options.Delay = 0
	}

	file, header, err := r.FormFile("attachment")
	if err != nil {
		response.Message = []any{
			"File recovery failed.",
			err.Error(),
		}
		return response
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		response.Message = []any{
			"Failed to read the file",
			err.Error(),
		}
	}

	body.Message.Filename = header.Filename
	body.Message.MediaType = r.FormValue("mediatype")
	body.Message.Caption = r.FormValue("caption")
	body.Message.GifPlayback = r.FormValue("isGif") == "true"
	body.Options.Group.HiddenMention = r.FormValue("groupHiddenMention") == "true"

	body.Message.Url = "none"

	mime := header.Header.Get("Content-Type")
	param := chi.URLParam(r, "instance")

	if strings.Contains(r.URL.Path, "/send/audio-file") {
		body.Message.MediaType = "audio"
		mime = "audio/ogg; codecs=opus"
		if body.Options.Presence != "recording" {
			body.Options.Presence = "none"
		}
	}

	if strings.Contains(r.URL.Path, "/send/ptv") {
		body.Message.MediaType = "ptv"
		if body.Options.Presence != "recording" {
			body.Options.Presence = "none"
		}
	}

	if errs := validate.Struct(body); errs != nil {
		response.Message = ErrorList(errs)
		return response
	}

	quoted, errs := validateQuoted(body.Options.QuotedMessage)
	if errs != nil {
		response.Message = ErrorList(errs)
		return response
	}

	data, status, err := h.service.FileMessage(param, &body, mime, bytes, quoted)

	response.StatusCode = status

	if err != nil {
		response.Message = []any{err.Error()}
		return response
	}

	response.Message = data
	return response
}

func (h *SendMessage) PostSendLocation(r *http.Request) *Response {
	var body contract.LocationMessage
	e := render.DecodeJSON(r.Body, &body)
	if err := UnmarshalDescriptionError(e); err != nil {
		return err
	}

	response := NewResponse(http.StatusBadRequest)

	if errs := validate.Struct(body); errs != nil {
		response.Message = ErrorList(errs)
		return response
	}

	quoted, errs := validateQuoted(body.Options.QuotedMessage)
	if errs != nil {
		response.Message = ErrorList(errs)
		return response
	}

	param := chi.URLParam(r, "instance")

	data, status, err := h.service.LocationMessage(param, &body, quoted)

	response.StatusCode = status

	if err != nil {
		response.Message = []any{err.Error()}
		return response
	}

	response.Message = data
	return response
}

func (h *SendMessage) PostSendContact(r *http.Request) *Response {
	var body contract.ContactMessage
	e := render.DecodeJSON(r.Body, &body)
	if err := UnmarshalDescriptionError(e); err != nil {
		return err
	}

	response := NewResponse(http.StatusBadRequest)

	if errs := validate.Struct(body); errs != nil {
		response.Message = ErrorList(errs)
		return response
	}

	quoted, errs := validateQuoted(body.Options.QuotedMessage)
	if errs != nil {
		response.Message = ErrorList(errs)
		return response
	}

	param := chi.URLParam(r, "instance")

	data, status, err := h.service.ContactMessage(param, &body, quoted)

	response.StatusCode = status

	if err != nil {
		response.Message = []any{err.Error()}
		return response
	}

	response.Message = data
	return response
}

// Deprecated: send list not work
func (h *SendMessage) PostSendList(r *http.Request) *Response {
	var body contract.ListMessage
	e := render.DecodeJSON(r.Body, &body)
	if err := UnmarshalDescriptionError(e); err != nil {
		return err
	}

	response := NewResponse(http.StatusBadRequest)

	if errs := validate.Struct(&body); errs != nil {
		response.Message = ErrorList(errs)
		return response
	}

	quoted, errs := validateQuoted(body.Options.QuotedMessage)
	if errs != nil {
		response.Message = ErrorList(errs)
		return response
	}

	param := chi.URLParam(r, "instance")

	data, status, err := h.service.ListMessage(param, &body, quoted)

	response.StatusCode = status

	if err != nil {
		response.Message = []any{err.Error()}
		return response
	}

	response.Message = data
	return response
}

func (h *SendMessage) PostSendPoll(r *http.Request) *Response {
	var body contract.PollMessage
	e := render.DecodeJSON(r.Body, &body)
	if err := UnmarshalDescriptionError(e); err != nil {
		return err
	}

	response := NewResponse(http.StatusBadRequest)

	if errs := validate.Struct(body); errs != nil {
		response.Message = ErrorList(errs)
		return response
	}

	quoted, errs := validateQuoted(body.Options.QuotedMessage)
	if errs != nil {
		response.Message = ErrorList(errs)
		return response
	}
	
	if body.Message.SelectableOptionsCount == 0 {
		body.Message.SelectableOptionsCount = 1
	}

	param := chi.URLParam(r, "instance")

	data, status, err := h.service.PoolMessage(param, &body, quoted)

	response.StatusCode = status

	if err != nil {
		response.Message = []any{err.Error()}
		return response
	}

	response.Message = data
	return response
}

func (h *SendMessage) PatchSendReaction(r *http.Request) *Response {
	var body contract.ReactionMessage
	e := render.DecodeJSON(r.Body, &body)
	if err := UnmarshalDescriptionError(e); err != nil {
		return err
	}

	response := NewResponse(http.StatusBadRequest)

	if errs := validate.Struct(body); errs != nil {
		response.Message = ErrorList(errs)
		return response
	}

	param := chi.URLParam(r, "instance")

	data, status, err := h.service.ReactionMessage(param, &body)

	response.StatusCode = status

	if err != nil {
		response.Message = []any{err.Error()}
		return response
	}

	response.Message = data
	return response
}

func (h *SendMessage) PatchEditMessage(r *http.Request) *Response {
	query := r.URL.Query()

	response := NewResponse(http.StatusBadRequest)

	messageId := query.Get("messageId")
	if messageId == "" {
		response.Message = ErrorList([]error{errors.New("Query 'messageId' empty or not defined.")})
		return response
	}

	var body contract.EditMessage
	e := render.DecodeJSON(r.Body, &body)
	if err := UnmarshalDescriptionError(e); err != nil {
		return err
	}

	if errs := validate.Struct(&body); errs != nil {
		response.Message = ErrorList(errs)
		return response
	}

	param := chi.URLParam(r, "instance")

	data, status, err := h.service.EditMessage(param, messageId, &body)

	response.StatusCode = status

	if err != nil {
		response.Message = []any{err.Error()}
		return response
	}

	response.Message = data
	return response
}
