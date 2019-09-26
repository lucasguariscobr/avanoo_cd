package utils

import (
	"log"
	"net/http"
	"time"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode   int
	responseSize int
}

func NewLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK, 0}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *loggingResponseWriter) Write(response []byte) (int, error) {
	lrw.responseSize += len(response)
	return lrw.ResponseWriter.Write(response)
}

func LoggingHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		t1 := time.Now()
		lw := NewLoggingResponseWriter(w)
		next.ServeHTTP(lw, r)
		t2 := time.Now()
		log.Printf("[%s] %q %v %v %v\n", r.Method, r.URL.String(), lw.statusCode, lw.responseSize, t2.Sub(t1))
	}

	return http.HandlerFunc(fn)
}
