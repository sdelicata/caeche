package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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

	mux.Handle("/songs/", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			rw.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprint(rw, "Method not allowed")
			return
		}

		time.Sleep(500 * time.Millisecond)
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
	}))

	mux.Handle("/api.json", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte("{}"))
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
