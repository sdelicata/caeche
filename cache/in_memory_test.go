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
			desc:      "Response too old for a request with Cache-Control: max-age=10",
			headers:   map[string]string{"Cache-Control": "max-age=10"},
			expiresIn: 60,
			expected:  false,
		},
		{
			desc:      "Response too old for a request with Cache-Control: s-max-age=10",
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

func TestStatusIsCacheable(t *testing.T) {
	testCases := map[int]bool{200: true, 201: false, 202: false, 203: true, 204: true, 205: false, 206: true, 207: false, 208: false, 226: false,
		300: true, 301: true, 302: false, 303: false, 304: false, 305: false, 306: false, 307: false, 308: false,
		400: false, 401: false, 402: false, 403: false, 404: true, 405: true, 406: false, 407: false, 408: false, 409: false, 410: true, 411: false, 412: false, 413: false, 414: true, 415: false, 416: false, 417: false, 418: false, 421: false, 422: false, 423: false, 424: false, 425: false, 426: false, 427: false, 428: false, 429: false, 431: false, 451: false,
		500: false, 501: true, 502: false, 503: false, 504: false, 505: false, 506: false, 507: false, 508: false, 509: false, 510: false, 511: false,
	}
	for status, cacheable := range testCases {
		var desc string
		if cacheable {
			desc = "%d is cacheable"
		} else {
			desc = "%d is not cacheable"
		}
		t.Run(desc, func(t *testing.T) {
			t.Parallel()
			res := http.Response{StatusCode: status}
			cache := NewInMemory(120)
			assert.Equal(t, cacheable, cache.IsCacheable(&res))
		})
	}
}

func TestIsCacheable(t *testing.T) {
	testCases := []struct {
		desc     string
		headers  map[string]string
		expected bool
	}{
		{
			desc:     "Response without Cache-Control header is cacheable",
			headers:  map[string]string{},
			expected: true,
		},
		{
			desc:     "Response with Cache-Control: max-age=3600 is cacheable",
			headers:  map[string]string{"Cache-Control": "max-age=3600"},
			expected: true,
		},
		{
			desc:     "Response with Cache-Control: no-cache isn't cacheable",
			headers:  map[string]string{"Cache-Control": "no-cache"},
			expected: false,
		},
		{
			desc:     "Response with Cache-Control: no-store isn't cacheable",
			headers:  map[string]string{"Cache-Control": "no-store"},
			expected: false,
		},
		{
			desc:     "Response with Cache-Control: max-age=0 isn't cacheable",
			headers:  map[string]string{"Cache-Control": "max-age=0"},
			expected: false,
		},
		{
			desc:     "Response with Cache-Control: s-max-age=0 isn't cacheable",
			headers:  map[string]string{"Cache-Control": "s-max-age=0"},
			expected: false,
		},
	}
	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			res := http.Response{
				StatusCode: http.StatusOK,
				Header:     map[string][]string{},
			}
			if len(test.headers) > 0 {
				for k, v := range test.headers {
					res.Header.Set(k, v)
				}
			}
			cache := NewInMemory(120)
			assert.Equal(t, test.expected, cache.IsCacheable(&res))
		})
	}
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
