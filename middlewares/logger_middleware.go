package middle

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

func LoggerMiddleware(logger *logrus.Entry) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			logger.WithFields(logrus.Fields{
				"method":  r.Method,
				"path":    r.URL.Path,
				"time":    time.Since(start),
			}).Info("request handled")

			next.ServeHTTP(w, r)
		})
	}
}
