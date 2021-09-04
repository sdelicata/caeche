package cache

import (
	"bytes"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Cache interface {
	Get(req *http.Request) (Response, bool)
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

func AcceptsCache(req *http.Request) bool {
	if !(req.Method == http.MethodGet || req.Method == http.MethodHead) ||
		req.Header.Get("Pragma") == "no-cache" ||
		strings.Contains(req.Header.Get("Cache-Control"), "no-cache") ||
		strings.Contains(req.Header.Get("Cache-Control"), "no-store") ||
		strings.Contains(req.Header.Get("Cache-Control"), "max-age=0") ||
		strings.Contains(req.Header.Get("Cache-Control"), "s-max-age=0") ||
		req.Header.Get("Authorization") != "" {
		log.Debugf("Request doesn't accept cache")
		return false
	} else {
		log.Debugf("Request accepts cache")
		return true
	}
}

func IsValidForRequest(response Response, req *http.Request) bool {
	if isTooOldForRequest(response, req) {
		log.Debugf("Response too old for the request")
		return false
	}
	if hasExpired(response) {
		log.Debugf("Response expired")
		return false
	}
	return true
}

func IsNotModified(req *http.Request) bool {
	return req.Header.Get("If-Modified-Since") != "" ||
		req.Header.Get("If-Unmodified-Since") != "" ||
		req.Header.Get("If-None-Match") != "" ||
		req.Header.Get("If-Match") != ""
}

func IsCacheable(res *http.Response) bool {
	if !isStatusCacheable(res.StatusCode) {
		log.Debugf("Response status code non cacheable")
		return false
	}
	if ttl, ok := getTTL(res.Header); ok && ttl == time.Duration(0) {
		log.Debugf("Response non cacheable")
		return false
	}
	return true
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

func isTooOldForRequest(response Response, req *http.Request) bool {
	ttl, ok := getTTL(req.Header)
	if ok {
		return response.Created.Before(time.Now().Add(ttl * -1))
	}
	return false
}

func hasExpired(response Response) bool {
	return response.Expires.Before(time.Now().UTC())
}

func isStatusCacheable(status int) bool {
	cacheableStatus := []int{200, 203, 204, 206, 300, 301, 404, 405, 410, 414, 501}
	for _, v := range cacheableStatus {
		if v == status {
			return true
		}
	}
	return false
}

func getTTL(headers http.Header) (time.Duration, bool) {
	expires := headers.Get("If-Modified-Since")
	if expires != "" {
		expiresDate, err := http.ParseTime(expires)
		if err == nil {
			diff := expiresDate.UTC().Sub(time.Now().UTC())
			if diff > 0 {
				return diff, true
			}
		}
	}

	cacheControl := headers.Get("Cache-Control")
	if cacheControl != "" {
		if strings.Contains(cacheControl, "no-cache") ||
			strings.Contains(cacheControl, "no-store") ||
			strings.Contains(cacheControl, "private") {
			return time.Duration(0), true
		}
		ageRegex := regexp.MustCompile(`s?-?max-age=(?P<TTL>\d+)`)
		age := ageRegex.FindStringSubmatch(cacheControl)
		if len(age) > 0 {
			ageTTL, err := strconv.ParseInt(age[1], 10, 64)
			if err == nil {
				return time.Duration(ageTTL) * time.Second, true
			}
		}
	}

	return time.Duration(0), false
}
