package server

import (
	"crypto/tls"
	"fmt"
	"github.com/sdelicata/caeche/config"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"time"
)

func NewReverseProxy(config config.Config) http.Handler {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	client := http.DefaultClient

	return http.HandlerFunc(func (rw http.ResponseWriter, req *http.Request) {
		start := time.Now()

		req.Host = config.Backend.Host
		req.URL.Host = config.Backend.Host
		req.URL.Scheme = config.Backend.Scheme
		req.RequestURI = ""

		res, err := client.Do(req)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(rw, err)
			log.Errorf("%s [%+v] %q %s://%s%s (%d) %+v",
				req.RemoteAddr,
				start.UTC(),
				req.Method,
				req.URL.Scheme, req.URL.Host, req.URL.Path,
				500,
				time.Since(start),
			)
			return
		}

		rw.WriteHeader(res.StatusCode)

		for name, values := range res.Header {
			for _, value := range values {
				rw.Header().Set(name, value)
			}
		}

		io.Copy(rw, res.Body)

		log.Debugf("%s [%+v] \"%s\" %s://%s%s (%d) %+v",
			req.RemoteAddr,
			start.UTC(),
			req.Method,
			req.URL.Scheme, req.URL.Host, req.URL.Path,
			res.StatusCode,
			time.Since(start),
		)
	})
}


