package cache

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAcceptsCache(t *testing.T) {
	testCases := []struct {
		desc       string
		method     string
		headers    map[string]string
		defaultTTL int
		expected   bool
	}{
		{
			desc:       "GET request without cache header",
			method:     http.MethodGet,
			headers:    map[string]string{},
			defaultTTL: 3600,
			expected:   true,
		}, {
			desc:       "HEAD request without cache header",
			method:     http.MethodHead,
			headers:    map[string]string{},
			defaultTTL: 3600,
			expected:   true,
		},
		{
			desc:       "OPTIONS request without cache header",
			method:     http.MethodOptions,
			headers:    map[string]string{},
			defaultTTL: 3600,
			expected:   false,
		},
		{
			desc:       "TRACE request without cache header",
			method:     http.MethodTrace,
			headers:    map[string]string{},
			defaultTTL: 3600,
			expected:   false,
		},
		{
			desc:       "POST request without cache header",
			method:     http.MethodPost,
			headers:    map[string]string{},
			defaultTTL: 3600,
			expected:   false,
		},
		{
			desc:       "PUT request without cache header",
			method:     http.MethodPut,
			headers:    map[string]string{},
			defaultTTL: 3600,
			expected:   false,
		},
		{
			desc:       "PATCH request without cache header",
			method:     http.MethodPatch,
			headers:    map[string]string{},
			defaultTTL: 3600,
			expected:   false,
		},
		{
			desc:       "DELETE request without cache header",
			method:     http.MethodDelete,
			headers:    map[string]string{},
			defaultTTL: 3600,
			expected:   false,
		},
		{
			desc:       "CONNECT request without cache header",
			method:     http.MethodConnect,
			headers:    map[string]string{},
			defaultTTL: 3600,
			expected:   false,
		},
		{
			desc:       "GET request with Pragma: no-cache",
			method:     http.MethodGet,
			headers:    map[string]string{"Pragma": "no-cache"},
			defaultTTL: 3600,
			expected:   false,
		},
		{
			desc:       "GET request with Cache-Control: no-cache",
			method:     http.MethodGet,
			headers:    map[string]string{"Cache-Control": "no-cache"},
			defaultTTL: 3600,
			expected:   false,
		},
		{
			desc:       "GET request with Cache-Control: no-store",
			method:     http.MethodGet,
			headers:    map[string]string{"Cache-Control": "no-store"},
			defaultTTL: 3600,
			expected:   false,
		},
		{
			desc:       "GET request with Cache-Control: max-age=0",
			method:     http.MethodGet,
			headers:    map[string]string{"Cache-Control": "max-age=0"},
			defaultTTL: 3600,
			expected:   false,
		},
		{
			desc:       "GET request with Cache-Control: max-age=60",
			method:     http.MethodGet,
			headers:    map[string]string{"Cache-Control": "max-age=60"},
			defaultTTL: 3600,
			expected:   true,
		},
		{
			desc:       "GET request with Cache-Control: s-max-age=0",
			method:     http.MethodGet,
			headers:    map[string]string{"Cache-Control": "s-max-age=0"},
			defaultTTL: 3600,
			expected:   false,
		},
		{
			desc:       "GET request with Cache-Control: s-max-age=60",
			method:     http.MethodGet,
			headers:    map[string]string{"Cache-Control": "s-max-age=60"},
			defaultTTL: 3600,
			expected:   true,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest(test.method, "http://localhost", nil)
			if len(test.headers) > 0 {
				for k, v := range test.headers {
					req.Header.Set(k, v)
				}
			}
			cache := NewInMemory(3600)
			assert.Equal(t, test.expected, cache.AcceptsCache(req))
		})
	}
}

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

func TestIsValidForRequest(t *testing.T) {
	testCases := []struct {
		desc      string
		headers   map[string]string
		expiresIn int
		expected  bool
	}{
		{
			desc:      "Valid response for a request with Cache-Control: max-age=3600",
			headers:   map[string]string{"Cache-Control": "max-age=3600"},
			expiresIn: 60,
			expected:  true,
		},
		{
			desc:      "Valid response for a request with Cache-Control: s-max-age=3600",
			headers:   map[string]string{"Cache-Control": "s-max-age=3600"},
			expiresIn: 60,
			expected:  true,
		},
		{
			desc:      "Response too old for a raquest with Cache-Control: max-age=10",
			headers:   map[string]string{"Cache-Control": "max-age=10"},
			expiresIn: 60,
			expected:  false,
		},
		{
			desc:      "Response too old for a raquest with Cache-Control: s-max-age=10",
			headers:   map[string]string{"Cache-Control": "s-max-age=10"},
			expiresIn: 60,
			expected:  false,
		},
		{
			desc:      "Response has expired",
			headers:   map[string]string{},
			expiresIn: 10 * -1,
			expected:  false,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			response := Response{
				Created: time.Now().Add(60 * -1 * time.Second),
				Expires: time.Now().Add(time.Duration(test.expiresIn) * time.Second),
			}
			req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
			if len(test.headers) > 0 {
				for k, v := range test.headers {
					req.Header.Set(k, v)
				}
			}
			cache := NewInMemory(120)
			assert.Equal(t, test.expected, cache.IsValidForRequest(response, req))
		})
	}
}
