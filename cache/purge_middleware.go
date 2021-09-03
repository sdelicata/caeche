package cache

import (
	log "github.com/sirupsen/logrus"
	"net/http"
)

func NewPurgeMiddleware(cache Cache) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {

			if req.Method != "PURGE" {
				next.ServeHTTP(rw, req)
				return
			}

			log.Debug("-----------------------")
			cache.Purge(req)
			rw.WriteHeader(http.StatusNoContent)
		})
	}
}
