package guards

import (
	"net/http"

	handler "codechat.dev/api/handlers"
	"codechat.dev/internal/whatsapp"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/sirupsen/logrus"
)

type AuthGuard struct {
	logger      *logrus.Entry
	globalToken string
	instance    *whatsapp.Instance
}

func NewAuthGuard(instance *whatsapp.Instance, globalToken string) *AuthGuard {
	logger := logrus.New()
	return &AuthGuard{
		logger:      logger.WithFields(logrus.Fields{"name": "auth-guard"}),
		globalToken: globalToken,
		instance:    instance,
	}
}

func (a *AuthGuard) CanActivate() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := handler.NewResponse(http.StatusUnauthorized)
			render.Status(r, response.StatusCode)

			param := chi.URLParam(r, "instance")

			if param != a.instance.Name {
				render.JSON(w, r, response.GetResponse())
				a.logger.WithFields(logrus.Fields{
					"param":        param,
					"instanceName": a.instance.Name,
				}).Error(response)
				return
			}

			apikey := r.Header.Get("apikey")

			adminGuard := NewAdminGuard(a.globalToken)

			if apikey == "" {
				render.JSON(w, r, response.GetResponse())
				a.logger.WithFields(logrus.Fields{
					"param":  param,
					"apikey": "empty",
				}).Error(response)
				return
			}

			activate := adminGuard.CanActivate(w, r)

			if activate != nil {
				if !activate.(bool) {
					return
				}
				next.ServeHTTP(w, r)
				return
			}

			if *a.instance.Apikey != apikey {
				render.JSON(w, r, response.GetResponse())
				a.logger.WithFields(logrus.Fields{
					"param":      param,
					"apikey":     apikey,
					"licenseKey": a.instance.Apikey,
				}).Error("invalid apikey")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
