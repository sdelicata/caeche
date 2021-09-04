package cache

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetNotFoundResponse(t *testing.T) {
	url := "http://localhost"
	req := httptest.NewRequest(http.MethodGet, url, nil)
	cache := NewInMemory(3600)
	response, ok := cache.Get(req)
	assert.Equal(t, Response{}, response)
	assert.Equal(t, false, ok)
}

func TestGetResponse(t *testing.T) {
	url := "http://localhost"
	req := httptest.NewRequest(http.MethodGet, url, nil)
	cache := NewInMemory(3600)
	key := cache.newStorageKeyFromRequest(req)
	expectedResponse := Response{
		URL:        url,
		Method:     http.MethodGet,
		StatusCode: http.StatusOK,
		Created:    time.Now().Add(3600 * -1 * time.Second),
		Expires:    time.Now().Add(3600 * time.Second),
	}
	store := map[StorageKey]Response{key: expectedResponse}
	cache.SetStore(store)
	response, ok := cache.Get(req)
	assert.Equal(t, expectedResponse, response)
	assert.Equal(t, true, ok)
}

func TestSave(t *testing.T) {
	cache := NewInMemory(3600)
	store := map[StorageKey]Response{}
	cache.SetStore(store)
	response := Response{
		URL:             "http://localhost",
		Method:          http.MethodGet,
		StatusCode:      http.StatusOK,
		ResponseHeaders: http.Header{},
		Body:            []byte{},
		Created:         time.Now(),
	}
	cache.Save(response)
	assert.Len(t, store, 1)
}

func TestSaveVariationsOfTheSameResource(t *testing.T) {
	cache := NewInMemory(3600)
	store := map[StorageKey]Response{}
	cache.SetStore(store)

	response := Response{
		URL:        "http://localhost",
		Method:     http.MethodGet,
		StatusCode: http.StatusOK,
		Created:    time.Now(),
	}
	cache.Save(response)

	responseWithQueryParams := Response{
		URL:        "http://localhost?foo=bar",
		Method:     http.MethodGet,
		StatusCode: http.StatusOK,
		Created:    time.Now(),
	}
	cache.Save(responseWithQueryParams)

	responseWithRequestHeaders := Response{
		URL:            "http://localhost",
		Method:         http.MethodGet,
		StatusCode:     http.StatusOK,
		RequestHeaders: http.Header{},
		Created:        time.Now(),
	}
	responseWithRequestHeaders.RequestHeaders.Set("X-Foo", "bar")
	cache.Save(responseWithRequestHeaders)
	assert.Len(t, store, 3)
}

func TestPurge(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
	cache := NewInMemory(3600)
	key := cache.newStorageKeyFromRequest(req)
	store := map[StorageKey]Response{key: {
		URL:        req.URL.String(),
		Method:     req.Method,
		StatusCode: http.StatusOK,
		Created:    time.Now().Add(3600 * -1 * time.Second),
		Expires:    time.Now().Add(3600 * time.Second),
	}}
	cache.SetStore(store)
	cache.Purge(req)
	assert.Len(t, store, 0)
}
