package server

import (
	"crypto/tls"
	"github.com/sdelicata/caeche/config"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

func NewReverseProxy(config config.Config) http.Handler {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	client := http.DefaultClient

	pool := NewCachePool()

	return http.HandlerFunc(func (rw http.ResponseWriter, req *http.Request) {
		start := time.Now()
		var cacheFlag string

		req.Host = config.Backend.Host
		req.URL.Host = config.Backend.Host
		req.URL.Scheme = config.Backend.Scheme
		req.RequestURI = ""
		remoteAddr, _, _ := net.SplitHostPort(req.RemoteAddr)

		if !IsRequestCacheable(req) {
			cacheFlag = "NOT CACHEABLE"
		} else {
			key := NewStorageKeyFromRequest(req)
			cachedResponse, ok := pool.Get(key)
			if !ok {
				cacheFlag = "MISS"
			} else {
				cacheFlag = "HIT"
				for name, values := range cachedResponse.ResponseHeaders {
					for _, value := range values {
						rw.Header().Set(name, value)
					}
				}
				rw.WriteHeader(cachedResponse.StatusCode)
				rw.Write(cachedResponse.Body)
				log.Infof("[%+v] %s \"%s\" %s://%s%s (%d) %+v [%s]",
					start.UTC(),
					remoteAddr,
					req.Method,
					req.URL.Scheme, req.URL.Host, req.URL.Path,
					cachedResponse.StatusCode,
					time.Since(start),
					cacheFlag,
				)
				return
			}
		}

		req.Header.Set("X-Forwarded-For", remoteAddr)

		res, err := client.Do(req)
		if err != nil {
			rw.WriteHeader(http.StatusBadGateway)
			log.Error(err)
			log.Infof("[%+v] %s %q %s://%s%s (%d) %+v",
				time.Now().UTC(),
				remoteAddr,
				req.Method,
				req.URL.Scheme, req.URL.Host, req.URL.Path,
				http.StatusBadGateway,
				time.Since(start),
			)
			return
		}
		defer res.Body.Close()

		for name, values := range res.Header {
			for _, value := range values {
				rw.Header().Set(name, value)
			}
		}
		rw.WriteHeader(res.StatusCode)

		body, err := ioutil.ReadAll(res.Body)

		if err != nil {
			log.Errorf("ERROR")
		}
		rw.Write(body)

		if IsRequestCacheable(req) && IsResponseCacheable(res) {
			response := Response{
				URL:             req.URL,
				Method:          req.Method,
				StatusCode:      res.StatusCode,
				ResponseHeaders: res.Header,
				Body:            body,
			}
			pool.Save(response)
		}

		log.Infof("[%+v] %s \"%s\" %s://%s%s (%d) %+v [%s]",
			start.UTC(),
			remoteAddr,
			req.Method,
			req.URL.Scheme, req.URL.Host, req.URL.Path,
			res.StatusCode,
			time.Since(start),
			cacheFlag,
		)
	})
}


