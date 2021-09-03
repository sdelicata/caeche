package cache

import (
	"bytes"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"time"
)

type Cache interface {
	AcceptsCache(req *http.Request) bool
	Get(req *http.Request) (Response, bool)
	IsValidForRequest(response Response, req *http.Request) bool
	IsCacheable(req *http.Response) bool
	Save(res Response)
	Purge(req *http.Request)
}

type Response struct {
	URL             string
	Method          string
	StatusCode      int
	RequestHeaders  http.Header
	ResponseHeaders http.Header
	Body            []byte
	Created         time.Time
	Expires         time.Time
}

func WriteResponse(rw http.ResponseWriter, response Response) {
	for name, values := range response.ResponseHeaders {
		for _, value := range values {
			rw.Header().Set(name, value)
		}
	}
	rw.WriteHeader(response.StatusCode)
	_, err := io.Copy(rw, io.NopCloser(bytes.NewBuffer(response.Body)))
	if err != nil {
		log.Fatal(err)
	}
}
