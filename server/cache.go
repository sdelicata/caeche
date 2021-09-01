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
	success := "MISS"
	res, ok := pool[key]
	if ok == true {
		success = "HIT"
	}
	log.Debugf("CACHE - Retreiving response {%+v} from cache pool [%s]", key, success)
	return res, ok
}

func IsRequestCacheable(req *http.Request) bool {
	switch {
	case !IsMethodSafe(req.Method) :
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