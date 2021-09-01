package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

func main() {
	mux := http.NewServeMux()

	mux.Handle("/", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/" {
			http.NotFound(rw, req)
			return
		}

		time.Sleep(500 * time.Millisecond)
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte("--- Request dump ---\n\n"))
		req.Write(rw)
	}))

	mux.Handle("/status", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			rw.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		url, _ := url.Parse(req.URL.String())
		queryParams := url.Query()
		status, _ := strconv.Atoi(queryParams.Get("status"))
		if status == http.StatusMovedPermanently {
			rw.Header().Set("Location", "/new-location")
		}
		rw.WriteHeader(status)
	}))

	mux.Handle("/api/songs/", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		switch {
		case req.Method == http.MethodPut :
		case req.Method == http.MethodPatch :
		case req.Method == http.MethodDelete :
			rw.WriteHeader(http.StatusNoContent)
		case req.Method == http.MethodPost :
			body, err := ioutil.ReadAll(req.Body)
			defer req.Body.Close()
			if err != nil {
				rw.WriteHeader(http.StatusBadRequest)
				return
			}
			if json.Valid(body) == false {
				rw.WriteHeader(http.StatusBadRequest)
				return
			}
			rw.Header().Set("Content-Type", "application/json")
			rw.WriteHeader(http.StatusCreated)
			rw.Write(body)
		default:
			rw.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))

	handler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		scheme := "http"
		if req.TLS != nil {
			scheme = "https"
		}
		fmt.Printf("%s://%s%s\n", scheme, req.Host, req.RequestURI)
		mux.ServeHTTP(rw, req)
	})

	http.ListenAndServe(":8000", handler)
}
