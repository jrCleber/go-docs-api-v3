package guards

import (
	"net/http"
	"slices"
	"strings"

	handler "codechat.dev/api/handlers"
	"codechat.dev/internal/domain/instance"
	"codechat.dev/internal/whatsapp"
	"codechat.dev/pkg/utils"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/sirupsen/logrus"
)

type InstanceGuard struct {
	instances   *instance.MapInstances
	store       *whatsapp.Store
	globalToken string
	logger      *logrus.Entry
}

func NewInstanceGuard(instances *instance.MapInstances, store *whatsapp.Store, globalToken string) *InstanceGuard {
	logger := logrus.New()
	return &InstanceGuard{
		instances:   instances,
		store:       store,
		globalToken: globalToken,
		logger:      logger.WithFields(logrus.Fields{"name": "instance-guard"}),
	}
}

func (i *InstanceGuard) IsAnInstance(w http.ResponseWriter, r *http.Request) bool {
	adminGuard := NewAdminGuard(i.globalToken)

	if activate := adminGuard.CanActivate(w, r); activate != nil {
		if !activate.(bool) {
			return !activate.(bool)
		}
		return true
	}

	param := chi.URLParam(r, "instance")
	response := handler.NewResponse(http.StatusBadRequest)

	findInstance, err := i.store.Read(param)

	render.Status(r, response.StatusCode)

	if err != nil {
		response.Message = []any{
			"Invalid instance: " + param,
			err.Error(),
		}
		render.JSON(w, r, response.GetResponse())
		return false
	}

	if slices.Contains(adminGuard.Allow.Paths, r.URL.Path) &&
		slices.Contains(adminGuard.Allow.Methods, r.Method) &&
		findInstance.Client == nil {
		return true
	}

	return true
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

	findInstance, err := i.store.Read(param)
	if err != nil {
		response.Message = []any{
			"Invalid instance: " + param,
			err.Error(),
		}
		render.JSON(w, r, response.GetResponse())
		return false
	}

	if isConnectRoute(r) {
		return true
	}

	instance := (*i.instances)[findInstance.ID]

	if instance.Client != nil && !instance.Client.IsLoggedIn() {
		response.StatusCode = http.StatusForbidden
		response.Message = []any{utils.StringJoin("", "Instance ", param, " not connected.")}

		render.Status(r, response.StatusCode)
		render.JSON(w, r, response.GetResponse())
		return false
	}

	return true
}
