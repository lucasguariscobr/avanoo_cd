package utils

import (
	"encoding/json"
	"log"
	"net/http"
)

func RecoverHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic: %+v", err)
				w.WriteHeader(http.StatusInternalServerError)
				jsonError := JSONErrror{"Error while processing your request, please retry later"}
				json.NewEncoder(w).Encode(jsonError)
			}
		}()

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
