package internalhttp

import (
	"fmt"
	"net/http"
	"time"
)

func loggingMiddleware(next http.Handler, logger Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)

		logger.Info(
			fmt.Sprintf(
				"%s [%s] %s %s %s (%.2fs)",
				r.RemoteAddr,
				time.Now().Format("2006-01-02 15:04:05"),
				r.Method,
				r.URL,
				r.UserAgent(),
				time.Since(start).Seconds(),
			),
		)
	})
}
