package main

import (
	"flag"
	"log"
	"net/http"
	"net/http/httputil"

	"github.com/lucas-clemente/quic-go/h2quic"
)

func main() {
	log.SetFlags(0)

	var addr, upstream string
	var keyFile, certFile string
	flag.StringVar(&addr, "addr", ":8080", "host:port listen address")
	flag.StringVar(&upstream, "upstream", "", "host:port of upstream server")
	flag.StringVar(&keyFile, "key", "", "TLS key file")
	flag.StringVar(&certFile, "cert", "", "TLS cert file")
	flag.Parse()

	if upstream == "" {
		log.Fatal("no upstream address")
	}

	log.Print("Starting quicd")
	log.Print("Forwarding to ", upstream)

	handle := func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.RequestURI)
		httputil.NewSingleHostReverseProxy(r.URL).ServeHTTP(w, r)
	}
	http.Handle("/", http.HandlerFunc(handle))

	go func() {
		log.Print("Listening HTTP on ", addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		log.Print("Listening QUIC on ", addr)
		if err := h2quic.ListenAndServeQUIC(addr, certFile, keyFile, nil); err != nil {
			log.Fatal(err)
		}
	}()

	log.Print("CTRL-C to exit")
	select {}
}
