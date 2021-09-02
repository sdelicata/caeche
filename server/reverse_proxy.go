package server

import (
	"bytes"
	"crypto/tls"
	cachePackage "github.com/sdelicata/caeche/cache"
	"github.com/sdelicata/caeche/config"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"net/http"
	"time"
)

type ReverseProxy http.Handler

func NewReverseProxy(config config.Config, cache cachePackage.Cache) http.Handler {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	client := http.DefaultClient

	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		log.Debug("-----------------------")
		start := time.Now()

		req.Host = config.Backend.Host
		req.URL.Host = config.Backend.Host
		req.URL.Scheme = config.Backend.Scheme
		req.RequestURI = ""
		remoteAddr, _, _ := net.SplitHostPort(req.RemoteAddr)

		if cache.AcceptsCache(req) {
			cachedResponse, ok := cache.Get(req)
			if ok && cache.IsValidForRequest(cachedResponse, req) {
				cachePackage.WriteResponse(rw, cachedResponse)
				logRequest(req, start, remoteAddr, cachedResponse.StatusCode, "HIT")
				return
			}
		}

		req.Header.Set("X-Forwarded-For", remoteAddr)

		log.Debugf("Fetching %s", req.URL)
		res, err := client.Do(req)
		if err != nil {
			rw.WriteHeader(http.StatusBadGateway)
			log.Error(err)
			log.Infof("[%+v] %s %q %s (%d) %+v",
				start.UTC(),
				remoteAddr,
				req.Method,
				req.URL,
				http.StatusBadGateway,
				time.Since(start),
			)
			return
		}
		defer func() {
			err := res.Body.Close()
			if err != nil {
				log.Fatal(err)
			}
		}()

		for name, values := range res.Header {
			for _, value := range values {
				rw.Header().Set(name, value)
			}
		}
		rw.WriteHeader(res.StatusCode)

		done := make(chan bool)
		go func() {
			for {
				select {
				case <-time.Tick(10 * time.Millisecond):
					rw.(http.Flusher).Flush()
				case <-done:
					return
				}
			}
		}()

		var buffer bytes.Buffer
		mrw := io.MultiWriter(rw, &buffer)
		_, err = io.Copy(mrw, res.Body)
		if err != nil {
			log.Fatal(err)
		}

		close(done)

		if cache.IsCacheable(res) {
			log.Debugf("%+v", buffer)
			cache.Save(cachePackage.Response{
				URL:             res.Request.URL.String(),
				Method:          res.Request.Method,
				StatusCode:      res.StatusCode,
				RequestHeaders:  res.Request.Header,
				ResponseHeaders: res.Header,
				Body:            buffer.Bytes(),
				Created:         start,
			})
		}

		logRequest(req, start, remoteAddr, res.StatusCode, "MISS")
	})
}

func logRequest(req *http.Request, start time.Time, remoteAddr string, statusCode int, flag string) {
	log.Infof("[%+v] %s \"%s\" %s (%d) %+v [%s]",
		start.UTC(),
		remoteAddr,
		req.Method,
		req.URL,
		statusCode,
		time.Since(start),
		flag,
	)
}
