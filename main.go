package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
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
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var b bytes.Buffer

		mw := io.MultiWriter(w, &b)

		fmt.Fprintf(mw, "Processing request\n")
		fmt.Fprintf(mw, "> Host: %v\n", r.Host)
		fmt.Fprintf(mw, "> URI: %v\n", r.URL)
		fmt.Fprintf(mw, "> Method: %v\n", r.Method)

		headers := make([]string, 0, len(r.Header))
		for k, _ := range r.Header {
			headers = append(headers, k)
		}
		sort.Strings(headers)
		for _, v := range headers {
			fmt.Fprintf(mw, "> %v: %v\n", v, r.Header.Get(v))
		}

		if body, err := ioutil.ReadAll(r.Body); err == nil {
			if len(body) > 0 {
				fmt.Fprintln(mw, "")
				fmt.Fprintln(mw, string(body))
			}
		} else {
			log.Println("Failed to read body:", err)
		}
		log.Println(b.String())
	})

	log.Println("Listening on", listen)
	log.Fatal(http.ListenAndServe(listen, nil))
}
