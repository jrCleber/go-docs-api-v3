package guards

import (
	"net/http"
	"slices"

	handler "codechat.dev/api/handlers"
	"github.com/go-chi/render"
	"github.com/sirupsen/logrus"
)

type required struct {
	Methods []string
	Paths   []string
}

type AdminGuard struct {
	logger      *logrus.Entry
	globalToken string
	Allow       required
}

func NewAdminGuard(globalToken string) *AdminGuard {
	logger := logrus.New()
	return &AdminGuard{
		logger:      logger.WithFields(logrus.Fields{"name": "admin-guard"}),
		globalToken: globalToken,
		Allow: required{
			Methods: []string{"POST", "GET"},
			Paths:   []string{"/api/v3/instance"},
		},
	}
}

func (a *AdminGuard) CanActivate(w http.ResponseWriter, r *http.Request) any {
	response := handler.NewResponse(http.StatusUnauthorized)
	apikey := r.Header.Get("apikey")

	var activate any

	if slices.Contains(a.Allow.Methods, r.Method) && slices.Contains(a.Allow.Paths, r.URL.Path) {
		if apikey != a.globalToken {
			render.JSON(w, r, response.GetResponse())
			a.logger.Error(response, " - msg: invalid global apikey")
			activate = false
		}
		activate = true
	}
	return activate
}
