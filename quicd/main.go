package main

import (
	"flag"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/lucas-clemente/quic-go/h2quic"
)

func main() {
	log.SetFlags(0)

	var addr, addrtls, upstream string
	var keyFile, certFile string
	flag.StringVar(&addr, "addr", ":8080", "host:port for HTTP")
	flag.StringVar(&addr, "addrtls", ":8443", "host:port for HTTPS")
	flag.StringVar(&upstream, "upstream", "", "http://host:port/ of upstream server")
	flag.StringVar(&keyFile, "key", "", "TLS key file")
	flag.StringVar(&certFile, "cert", "", "TLS cert file")
	flag.Parse()

	if upstream == "" {
		log.Fatal("no upstream address")
	}

	upstreamURL, err := url.Parse(upstream)
	if err != nil {
		log.Fatal("Invalid upstream URL: ", err)
	}

	log.Print("Starting quicd")
	log.Print("Forwarding to ", upstreamURL)

	http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.RequestURI)
		httputil.NewSingleHostReverseProxy(upstreamURL).ServeHTTP(w, r)
	}))

	go func() {
		if addr == "" {
			return
		}
		log.Print("Listening HTTP on ", addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		if addrtls == "" {
			return
		}
		log.Print("Listening HTTPS on ", addrtls)
		if err := http.ListenAndServeTLS(addrtls, certFile, keyFile, nil); err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		if addr == "" {
			return
		}
		log.Print("Listening QUIC on ", addr)
		if err := h2quic.ListenAndServeQUIC(addr, certFile, keyFile, nil); err != nil {
			log.Fatal(err)
		}
	}()

	log.Print("CTRL-C to exit")
	select {}
}
