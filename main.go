package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var help bool
var listen string

func init() {
	flag.BoolVar(&help, "help", false, "Help message")
	flag.StringVar(&listen, "listen-address", ":8080", "Listen address")
}

func main() {
	flag.Parse()
	if help {
		fmt.Fprintln(os.Stderr, "Simple HTTP server displaying the incoming request")
		flag.PrintDefaults()
		os.Exit(0)
	}
	http.HandleFunc("/metrics", promhttp.Handler().ServeHTTP)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var b bytes.Buffer

		mw := io.MultiWriter(w, &b)

		fmt.Fprintf(mw, "Processing request\n")
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
			time.Sleep(d)
		}
	})

	log.Println("Listening on", listen)
	log.Fatal(http.ListenAndServe(listen, nil))
}
