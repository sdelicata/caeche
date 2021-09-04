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

	mux.Handle("/dump", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		time.Sleep(500 * time.Millisecond)
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte("--- Request dump ---\n\n"))
		req.Write(rw)
	}))

	mux.Handle("/cache-control", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			rw.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		url, _ := url.Parse(req.URL.String())
		queryParams := url.Query()
		val := queryParams.Get("val")
		rw.Header().Set("Cache-Control", val)
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

	mux.Handle("/stream", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte("Here comes \n"))
		rw.(http.Flusher).Flush()
		time.Sleep(2 * time.Second)
		rw.Write([]byte("pieces of my \n"))
		rw.(http.Flusher).Flush()
		time.Sleep(2 * time.Second)
		rw.Write([]byte("streamed content"))
		rw.(http.Flusher).Flush()
	}))

	handler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		scheme := "http"
		if req.TLS != nil {
			scheme = "https"
		}
		fmt.Printf("%s://%s%s\n", scheme, req.Host, req.RequestURI)
		mux.ServeHTTP(rw, req)
	})

	go http.ListenAndServe(":8000", handler)
	http.ListenAndServeTLS(":4443", "./cert.pem", "./key.pem", handler)
}
