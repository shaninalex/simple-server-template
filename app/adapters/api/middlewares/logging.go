package middlewares

import (
	"fmt"
	"net/http"
	"time"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		fmt.Printf(
			"%s %s (%fs)\n",
			time.Now().Format(time.RFC3339),
			r.Pattern,
			time.Since(start).Seconds(),
		)
	})
}
