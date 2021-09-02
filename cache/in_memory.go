package cache

import (
	log "github.com/sirupsen/logrus"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type StorageKey string

type InMemory struct {
	DefaultTTL int
	store      map[StorageKey]Response
}

func NewCacheInMemory(defaultTTL int) InMemory {
	return InMemory{
		DefaultTTL: defaultTTL,
		store:      make(map[StorageKey]Response),
	}
}

func (cache InMemory) AcceptsCache(req *http.Request) bool {
	switch {
	case !(req.Method == http.MethodGet || req.Method == http.MethodHead) :
		return false
	case strings.Contains(req.Header.Get("Cache-Control"), "no-cache") :
		return false
	case strings.Contains(req.Header.Get("Cache-Control"), "no-store") :
		return false
	case req.Header.Get("Pragma") == "no-cache" :
		return false
	default :
		return true
	}
}

func (cache InMemory) Get(req *http.Request) (Response, bool) {
	key := cache.newStorageKeyFromRequest(req)
	response, ok := cache.store[key]
	if !ok {
		log.Debugf("CACHE - Response {%+v} not found", key)
		return response, ok
	}
	if cache.isTooOldForRequest(response, req) {
		log.Debugf("CACHE - Response {%+v} too old for the request", key)
		return response, false
	}
	if cache.hasExpired(response) {
		log.Debugf("CACHE - Response {%+v} expired", key)
		return response, false
	}
	log.Debugf("CACHE - Response {%+v} retrieved", key)
	return response, ok
}

func (cache InMemory) IsCacheable(res *http.Response) bool {
	switch {
	case !cache.isStatusCacheable(res.StatusCode):
		return false
		//case cache.getTTL(res.Header) == time.Duration(0) :
		//	return false
	}
	return true
}

func (cache InMemory) Save(response Response) {
	key := cache.newStorageKeyFromResponse(response)
	ttl, ok := cache.getTTL(response.ResponseHeaders)
	if !ok {
		ttl = time.Duration(cache.DefaultTTL) * time.Second
	}
	response.Expires = response.Created.Add(ttl)
	cache.store[key] = response
	log.Debugf("CACHE - Response {%+v} saved", key)
}

func (cache InMemory) newStorageKeyFromRequest(req *http.Request) StorageKey {
	return StorageKey(req.Method + "_" + req.URL.String())
}

func (cache InMemory) newStorageKeyFromResponse(response Response) StorageKey {
	return StorageKey(response.Method + "_" + response.URL.String())
}

func (cache InMemory) isTooOldForRequest(response Response, req *http.Request) bool {
	ttl, ok := cache.getTTL(req.Header)
	if ok {
		return response.Created.Before(time.Now().Add(ttl * -1))
	}
	return false
}

func (cache InMemory) hasExpired(response Response) bool {
	return response.Expires.Before(time.Now())
}

func (cache InMemory) isStatusCacheable(status int) bool {
	cacheableStatus := []int{200, 203, 204, 206, 300, 301, 404, 405, 410, 414, 501}
	for _, v := range cacheableStatus {
		if v == status {
			return true
		}
	}
	return false
}

func (cache InMemory) getTTL(headers http.Header) (time.Duration, bool) {
	expires := headers.Get("Expires")
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
		if strings.Contains(cacheControl, "no-cache") || strings.Contains(cacheControl, "no-store") {
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