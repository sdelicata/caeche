package server

import (
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/url"
)

type Response struct {
	URL             *url.URL
	Method          string
	StatusCode      int
	RequestHeaders  http.Header
	ResponseHeaders http.Header
	Body            []byte
}

type StorageKey string

func NewStorageKeyFromRequest(req *http.Request) StorageKey {
	return StorageKey(req.Method + "_" + req.URL.String())
}

type CachePool map[StorageKey]Response

func NewCachePool() CachePool {
	return CachePool{}
}

func (pool CachePool) Save(res Response) {
	key := StorageKey(res.Method + "_" + res.URL.String())
	pool[key] = res
	log.Debugf("CACHE - Saving response {%+v} in cache pool", key)
}

func (pool CachePool) Get(key StorageKey) (Response, bool) {
	res, ok := pool[key]
	log.Debugf("CACHE - Retreiving response {%+v} from cache pool", key)
	return res, ok
}

func IsRequestCacheable(req *http.Request) bool {
	switch {
	case !IsMethodSafe(req.Method) :
		return false
	case IsNoCacheRequest(req) :
		return false
	}
	return true
}

func IsMethodSafe(method string) bool {
	safeMethods := []string{http.MethodGet, http.MethodHead}
	for _, v := range safeMethods {
		if v == method {
			return true
		}
	}
	return false
}

func IsNoCacheRequest(req *http.Request) bool {
	switch {
	case req.Header.Get("Cache-Control") == "no-cache" :
		return true
	case req.Header.Get("Cache-Control") == "no-store" :
		return true
	case req.Header.Get("Pragma") == "no-cache" :
		return true
	}
	return false
}

func IsResponseCacheable(res *http.Response) bool {
	switch {
	case !IsStatusCacheable(res.StatusCode):
		return false
	}
	return true
}

func IsStatusCacheable(status int) bool {
	cacheableStatus := []int{200, 203, 204, 206, 300, 301, 404, 405, 410, 414, 501}
	for _, v := range cacheableStatus {
		if v == status {
			return true
		}
	}
	return false
}