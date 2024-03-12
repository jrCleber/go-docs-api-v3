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
	store       *whatsapp.Store
	logger      *logrus.Entry
	globalToken string
}

func NewAuthGuard(store *whatsapp.Store, globalToken string) *AuthGuard {
	logger := logrus.New()
	return &AuthGuard{
		logger:      logger.WithFields(logrus.Fields{"name": "auth-guard"}),
		store:       store,
		globalToken: globalToken,
	}
}

func (a *AuthGuard) CanActivate() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := handler.NewResponse(http.StatusUnauthorized)
			render.Status(r, response.StatusCode)

			param := chi.URLParam(r, "instance")

			// auth := r.Header.Get("Authorization")

			// if auth == "" {
			// 	render.JSON(w, r, response.GetResponse())
			// 	return
			// }

			// token := strings.Replace(auth, "Bearer ", "", 1)

			// decode, err := a.Jwt.Read(&token)

			// if err != nil || !a.Jwt.IsScope(decode.RoleAccess) || !a.Jwt.VerifyAttributes(decode) {
			// 	render.JSON(w, r, response.GetResponse())
			// 	return
			// }

			apikey := r.Header.Get("apikey")

			adminGuard := NewAdminGuard(a.globalToken)

			if apikey == "" {
				render.JSON(w, r, response.GetResponse())
				a.logger.Error(response, apikey, " msg - empty apikey")
				return
			}

			activate := adminGuard.CanActivate(w, r)

			if  activate != nil {
				if !activate.(bool) {
					return
				}
				next.ServeHTTP(w, r)
				return
			}

			instance, err := a.store.Read(param)
			if err != nil {
				response.StatusCode = http.StatusBadRequest
				response.Message = []any{"Unable to read instance.", err.Error()}
				render.JSON(w, r, response.GetResponse())
				a.logger.Error(response)
				return
			}

			if instance == nil {
				response.Message = []any{
					"Invalid instance: " + param,
					"Instance not found",
				}
				render.JSON(w, r, response.GetResponse())
				a.logger.Error(response)
				return
			}

			if *instance.Apikey != apikey {
				render.JSON(w, r, response.GetResponse())
				a.logger.Error(response, " msg - invalid apikey")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
