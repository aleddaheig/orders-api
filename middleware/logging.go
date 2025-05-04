package middleware

import (
	"log"
	"net/http"
	"time"
)

// wrappedWriter is a custom response writer that tracks the status code
type wrappedWriter struct {
	http.ResponseWriter
	statusCode int
}

// Override WriteHeader to track the status code
func (w *wrappedWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.statusCode = statusCode
}

// Logging middleware logs the request method, URL, and duration
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wrapped := &wrappedWriter{w, http.StatusOK}
		next.ServeHTTP(wrapped, r)

		log.Printf("%d %s %s %s", wrapped.statusCode, r.Method, r.URL, time.Since(start))
	})
}
