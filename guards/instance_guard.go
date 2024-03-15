package guards

import (
	"net/http"
	"strings"

	handler "codechat.dev/api/handlers"
	"codechat.dev/internal/whatsapp"
	"codechat.dev/pkg/utils"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/sirupsen/logrus"
)

type InstanceGuard struct {
	instance    *whatsapp.Instance
	globalToken string
	logger      *logrus.Entry
}

func NewInstanceGuard(instance *whatsapp.Instance, globalToken string) *InstanceGuard {
	logger := logrus.New()
	return &InstanceGuard{
		instance:    instance,
		globalToken: globalToken,
		logger:      logger.WithFields(logrus.Fields{"name": "instance-guard"}),
	}
}

func isConnectRoute(r *http.Request) bool {
	path := r.URL.Path
	method := r.Method

	if method == "GET" && strings.Contains(path, "whatsapp/connect") {
		return true
	}

	return false
}

func (i *InstanceGuard) IsLoggedIn(w http.ResponseWriter, r *http.Request) bool {
	param := chi.URLParam(r, "instance")
	response := handler.NewResponse(http.StatusBadRequest)

	if param != i.instance.Name {
		render.JSON(w, r, response.GetResponse())
		i.logger.WithFields(logrus.Fields{
			"param":        param,
			"instanceName": i.instance.Name,
		}).Error(response)
		return false
	}

	if isConnectRoute(r) {
		return true
	}

	if i.instance.Client != nil && !i.instance.Client.IsLoggedIn() {
		response.StatusCode = http.StatusForbidden
		response.Message = []any{utils.StringJoin("", "Instance ", param, " not connected.")}

		render.Status(r, response.StatusCode)
		render.JSON(w, r, response.GetResponse())
		return false
	}

	return true
}
