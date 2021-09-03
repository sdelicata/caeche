package server

import (
	"bytes"
	cachePackage "github.com/sdelicata/caeche/cache"
	"github.com/sdelicata/caeche/config"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"net/http"
	"time"
)

type ReverseProxy struct {
	config config.Config
	cache  cachePackage.Cache
}

func NewReverseProxy(config config.Config, cache cachePackage.Cache) *ReverseProxy {
	return &ReverseProxy{
		config: config,
		cache:  cache,
	}
}

func (reverseProxy *ReverseProxy) GetHandler() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		log.Debug("-----------------------")
		start := time.Now()
		var cacheHit bool
		var cachedResponse cachePackage.Response

		// Prepare request to forward
		req.Host = reverseProxy.config.Backend.Host
		req.URL.Host = reverseProxy.config.Backend.Host
		req.URL.Scheme = reverseProxy.config.Backend.Scheme
		req.RequestURI = ""

		// Serve cache when it's possible
		acceptCache := reverseProxy.cache.AcceptsCache(req)
		if acceptCache {
			cachedResponse, cacheHit = reverseProxy.cache.Get(req)
			if cacheHit && reverseProxy.cache.IsValidForRequest(cachedResponse, req) {
				cachePackage.WriteResponse(rw, cachedResponse)
				logRequest(req, start, cachedResponse.StatusCode, "HIT")
				return
			}
		}

		// If not, forward the request to the backend
		res, err := reverseProxy.fetch(req)
		defer func() {
			err := res.Body.Close()
			if err != nil {
				log.Fatal(err)
			}
		}()

		// Error while fetching from backend: serve stale cache or 502
		if err != nil {
			log.Error(err)
			if cacheHit {
				log.Debug("Serving stale response")
				cachePackage.WriteResponse(rw, cachedResponse)
				logRequest(req, start, cachedResponse.StatusCode, "HIT")
				return
			}
			rw.WriteHeader(http.StatusBadGateway)
			log.Infof("[%+v] %q %s (%d) %+v",
				start.UTC(),
				req.Method,
				req.URL,
				http.StatusBadGateway,
				time.Since(start),
			)
			return
		}

		// Serve fetched response
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

		// Save cache if the response is cacheable
		if acceptCache && reverseProxy.cache.IsCacheable(res) {
			reverseProxy.cache.Save(cachePackage.Response{
				URL:             res.Request.URL.String(),
				Method:          res.Request.Method,
				StatusCode:      res.StatusCode,
				RequestHeaders:  res.Request.Header,
				ResponseHeaders: res.Header,
				Body:            buffer.Bytes(),
				Created:         start,
			})
		}

		logRequest(req, start, res.StatusCode, "MISS")
	})
}

func (reverseProxy *ReverseProxy) fetch(req *http.Request) (*http.Response, error) {
	log.Debugf("Fetching %s", req.URL)
	remoteAddr, _, _ := net.SplitHostPort(req.RemoteAddr)
	req.Header.Set("X-Forwarded-For", remoteAddr)
	res, err := http.DefaultClient.Do(req)
	delete(req.Header, "X-Forwarded-For")
	if err != nil {
		return nil, err
	}
	return res, nil
}

func logRequest(req *http.Request, start time.Time, statusCode int, flag string) {
	log.Infof("[%+v] \"%s\" %s (%d) %+v [%s]",
		start.UTC(),
		req.Method,
		req.URL,
		statusCode,
		time.Since(start),
		flag,
	)
}
