package main

import (
	"flag"
	"log"
	"net/http"
	"net/http/httputil"
)

func main() {
	log.SetFlags(0)

	var addr, upstream string
	flag.StringVar(&addr, "addr", ":8080", "host:port listen address")
	flag.StringVar(&upstream, "upstream", "", "host:port of upstream server")
	flag.Parse()

	if upstream == "" {
		log.Fatal("no upstream address")
	}

	http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.RequestURI)
		httputil.NewSingleHostReverseProxy(r.URL).ServeHTTP(w, r)
	}))

	log.Print("Listening on ", addr)
	log.Print("Forwarding to ", upstream)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
