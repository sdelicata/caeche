package cache

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
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

func NewInMemory(defaultTTL int) *InMemory {
	return &InMemory{
		DefaultTTL: defaultTTL,
		store:      make(map[StorageKey]Response),
	}
}

func (cache *InMemory) SetStore(store map[StorageKey]Response) {
	cache.store = store
}

func (cache *InMemory) AcceptsCache(req *http.Request) bool {
	if req.Header.Get("Pragma") == "no-cache" ||
		!(req.Method == http.MethodGet || req.Method == http.MethodHead) ||
		strings.Contains(req.Header.Get("Cache-Control"), "no-cache") ||
		strings.Contains(req.Header.Get("Cache-Control"), "no-store") ||
		strings.Contains(req.Header.Get("Cache-Control"), "max-age=0") ||
		strings.Contains(req.Header.Get("Cache-Control"), "s-max-age=0") {
		log.Debugf("Request doesn't accept cache")
		return false
	} else {
		log.Debugf("Request accepts cache")
		return true
	}
}

func (cache *InMemory) Get(req *http.Request) (Response, bool) {
	key := cache.newStorageKeyFromRequest(req)
	response, ok := cache.store[key]
	if !ok {
		log.Debugf("Getting %q : Response not found", key)
		return response, ok
	}
	log.Debugf("Getting %q : Response retrieved", key)
	return response, ok
}

func (cache *InMemory) IsValidForRequest(response Response, req *http.Request) bool {
	if cache.isTooOldForRequest(response, req) {
		log.Debugf("Response too old for the request")
		return false
	}
	if cache.hasExpired(response) {
		log.Debugf("Response expired")
		return false
	}
	return true
}

func (cache *InMemory) IsCacheable(res *http.Response) bool {
	if !cache.isStatusCacheable(res.StatusCode) {
		log.Debugf("Response status code non cacheable")
		return false
	}
	if ttl, ok := cache.getTTL(res.Header); ok && ttl == time.Duration(0) {
		log.Debugf("Response non cacheable")
		return false
	}
	return true
}

func (cache *InMemory) Save(response Response) {
	key := cache.newStorageKeyFromResponse(response)
	ttl, ok := cache.getTTL(response.ResponseHeaders)
	if !ok {
		ttl = time.Duration(cache.DefaultTTL) * time.Second
	}
	response.Expires = response.Created.Add(ttl)
	cache.store[key] = response
	log.Debugf("Saving %q : Response saved", key)
}

func (cache *InMemory) newStorageKeyFromRequest(req *http.Request) StorageKey {
	return StorageKey(fmt.Sprintf("%s_%s_%s", req.Method, req.URL, cache.hashHeaders(req.Header)))
}

func (cache *InMemory) newStorageKeyFromResponse(response Response) StorageKey {
	headers := response.RequestHeaders
	headers.Del("X-Forwarded-For")
	return StorageKey(fmt.Sprintf("%s_%s_%s", response.Method, response.URL, cache.hashHeaders(headers)))
}

func (cache *InMemory) hashHeaders(headers http.Header) string {
	jsonData, err := json.Marshal(headers)
	if err != nil {
		log.Error(err)
		return ""
	}
	hash := sha256.New()
	_, err = hash.Write(jsonData)
	if err != nil {
		log.Error(err)
		return ""
	}
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func (cache *InMemory) isTooOldForRequest(response Response, req *http.Request) bool {
	ttl, ok := cache.getTTL(req.Header)
	if ok {
		return response.Created.Before(time.Now().Add(ttl * -1))
	}
	return false
}

func (cache *InMemory) hasExpired(response Response) bool {
	return response.Expires.Before(time.Now())
}

func (cache *InMemory) isStatusCacheable(status int) bool {
	cacheableStatus := []int{200, 203, 204, 206, 300, 301, 404, 405, 410, 414, 501}
	for _, v := range cacheableStatus {
		if v == status {
			return true
		}
	}
	return false
}

func (cache *InMemory) getTTL(headers http.Header) (time.Duration, bool) {
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
