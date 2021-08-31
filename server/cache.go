package server

import (
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/url"
)

type Response struct {
	URL             url.URL
	Method          string
	StatusCode      int
	RequestHeaders  http.Header
	ResponseHeaders http.Header
	Content         [][]byte
}

func NewCacheMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Println("Executing CacheMiddleware")
			next.ServeHTTP(w, r)
			log.Println("Executing CacheMiddleware")
		})
	}
}