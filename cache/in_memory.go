package cache

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
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

func (cache *InMemory) Get(req *http.Request) (Response, bool) {
	key := cache.newStorageKeyFromRequest(req)
	response, ok := cache.store[key]
	if !ok {
		log.Debugf("Getting %q : Response not found in cache", key)
		return response, ok
	}
	log.Debugf("Getting %q : Response retrieved", key)
	return response, ok
}

func (cache *InMemory) Save(response Response) {
	key := cache.newStorageKeyFromResponse(response)
	ttl, ok := getTTL(response.ResponseHeaders)
	if !ok {
		ttl = time.Duration(cache.DefaultTTL) * time.Second
	}
	response.Expires = response.Created.Add(ttl)
	cache.store[key] = response
	log.Debugf("Saving %q : Response saved for %s", key, ttl)
}

func (cache *InMemory) Purge(req *http.Request) {
	pattern := fmt.Sprintf("%s_%s", req.URL, cache.hashHeaders(req.Header))
	for key := range cache.store {
		if strings.Contains(string(key), pattern) {
			delete(cache.store, key)
			log.Debugf("Purging %s", key)
		}
	}
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
