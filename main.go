package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var help bool
var listen string
var path string
var status int
var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))

func init() {
	flag.BoolVar(&help, "help", false, "Help message")
	flag.StringVar(&listen, "listen-address", ":8080", "Listen address")
	flag.IntVar(&status, "http-status", 200, "HTTP status response code")
	flag.StringVar(&path, "path", "/", "URL path")
}

func main() {
	flag.Parse()
	if help {
		fmt.Fprintln(os.Stderr, "Simple HTTP server displaying the incoming request")
		flag.PrintDefaults()
		os.Exit(0)
	}
	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		var b bytes.Buffer

		w.WriteHeader(status)
		mw := io.MultiWriter(w, &b)

		fmt.Fprintf(mw, "Processing request from %s\n", r.RemoteAddr)
		fmt.Fprintf(mw, "> Host: %v\n", r.Host)
		fmt.Fprintf(mw, "> URI: %v\n", r.URL)
		fmt.Fprintf(mw, "> Method: %v\n", r.Method)

		headers := make([]string, 0, len(r.Header))
		for k := range r.Header {
			headers = append(headers, k)
		}
		sort.Strings(headers)
		for _, v := range headers {
			fmt.Fprintf(mw, "> %v: %v\n", v, r.Header.Get(v))
		}

		if body, err := ioutil.ReadAll(r.Body); err == nil {
			if len(body) > 0 {
				fmt.Fprintln(mw, "")
				if r.Header.Get("Content-Type") == "application/json" {
					var o bytes.Buffer
					json.Indent(&o, body, "", "  ")
					fmt.Fprintln(mw, o.String())
				} else {
					fmt.Fprintln(mw, string(body))
				}
			}
		} else {
			log.Println("Failed to read body:", err)
		}
		log.Println(b.String())

		// Wait an optional time before returning to the client.
		q := r.URL.Query()
		if d, err := time.ParseDuration(q.Get("sleep")); err == nil {
			if _, ok := q["random"]; ok {
				d = time.Duration(float64(d) * rnd.Float64())
			}
			log.Printf("Sleeping for %s", d)
			time.Sleep(d)
		}
	})
	if path != "/metrics" {
		http.HandleFunc("/metrics", promhttp.Handler().ServeHTTP)
	}

	log.Printf("Listening on %s, path: %s\n", listen, path)
	log.Fatal(http.ListenAndServe(listen, nil))
}
