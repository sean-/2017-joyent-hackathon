package main

import (
	"flag"
	stdlog "log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	"github.com/lucas-clemente/quic-go/h2quic"
	"github.com/sean-/seed"
)

var logger *stdlog.Logger

func main() {
	log := stdlog.New(os.Stdout, "", stdlog.LstdFlags)
	log.SetFlags(0)
	logger = log

	var addr, addrtls, upstream string
	var keyFile, certFile string
	var circonusAPIKey string

	flag.StringVar(&addr, "addr", ":8080", "host:port for HTTP")
	flag.StringVar(&addrtls, "addrtls", ":8443", "host:port for HTTPS")
	flag.StringVar(&upstream, "upstream", "", "http://host:port/ of upstream server")
	flag.StringVar(&keyFile, "key", "", "TLS key file")
	flag.StringVar(&certFile, "cert", "", "TLS cert file")
	flag.StringVar(&circonusAPIKey, "circonus-api-key", "", "Circonus API Key")

	flag.Parse()

	if upstream == "" {
		log.Fatal("no upstream address")
	}

	upstreamURL, err := url.Parse(upstream)
	if err != nil {
		log.Fatal("Invalid upstream URL: ", err)
	}

	if err := initMetrics(circonusAPIKey); err != nil {
		log.Fatalf("Unable to initialize metrics: %v", err)
		os.Exit(1)
	}

	log.Print("Starting quicd")

	http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		const requestLatency = "request-latency"
		defer func() {
			metrics.Timing(requestLatency, float64(time.Now().Sub(start)))
		}()

		log.Printf("%s %s", r.Method, r.RequestURI)
		r.Host = upstreamURL.Host
		httputil.NewSingleHostReverseProxy(upstreamURL).ServeHTTP(w, r)
	}))

	go func() {
		if addr == "" {
			log.Print("HTTP disabled")
			return
		}
		log.Print("HTTP enabled on ", addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		if addrtls == "" || certFile == "" || keyFile == "" {
			log.Print("HTTPS disabled")
			return
		}
		log.Print("HTTPS enabled on ", addrtls)
		if err := http.ListenAndServeTLS(addrtls, certFile, keyFile, nil); err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		if addr == "" || certFile == "" || keyFile == "" {
			log.Print("QUIC disabled")
			return
		}
		log.Print("QUIC enabled on ", addr)
		if err := h2quic.ListenAndServeQUIC(addr, certFile, keyFile, nil); err != nil {
			log.Fatal(err)
		}
	}()

	time.Sleep(100 * time.Millisecond)
	log.Print("Forwarding to ", upstreamURL)
	log.Print("CTRL-C to exit")
	select {}
}

func init() {
	seed.MustInit()
}
