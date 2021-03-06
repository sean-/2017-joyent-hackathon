package main

import (
	"flag"
	stdlog "log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	cgm "github.com/circonus-labs/circonus-gometrics"
	"github.com/lucas-clemente/quic-go/h2quic"
	"github.com/pkg/errors"
	"github.com/sean-/seed"
)

var logger *stdlog.Logger
var metrics *cgm.CirconusMetrics

func main() {
	log := stdlog.New(os.Stdout, "", stdlog.LstdFlags)
	log.SetFlags(0)
	logger = log

	var httpAddr, httpsAddr, quicAddr, upstream string
	var keyFile, certFile string
	var circonusAPIKey string

	flag.StringVar(&httpAddr, "http", ":8080", "host:port for HTTP")
	flag.StringVar(&httpsAddr, "https", ":8443", "host:port for HTTPS")
	flag.StringVar(&quicAddr, "quic", ":8443", "host:port for QUIC")
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

	http.HandleFunc("/", metrics.TrackHTTPLatency("/", func(w http.ResponseWriter, r *http.Request) {
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
		if httpAddr == "" {
			log.Print("HTTP disabled")
			return
		}

		log.Print("HTTP enabled on ", httpAddr)
		if err := http.ListenAndServe(httpAddr, http.DefaultServeMux); err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		if httpsAddr == "" || certFile == "" || keyFile == "" {
			log.Print("HTTPS disabled")
			return
		}
		log.Print("HTTPS enabled on ", httpsAddr)
		if err := http.ListenAndServeTLS(httpsAddr, certFile, keyFile, http.DefaultServeMux); err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		if quicAddr == "" || certFile == "" || keyFile == "" {
			log.Print("QUIC disabled")
			return
		}
		log.Print("QUIC enabled on ", quicAddr)
		if err := h2quic.ListenAndServeQUIC(quicAddr, certFile, keyFile, http.DefaultServeMux); err != nil {
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

func initMetrics(circonusAPIKey string) (err error) {
	log := logger
	cmc := &cgm.Config{}
	cmc.Debug = false // set to true for debug messages
	cmc.Log = log

	cmc.CheckManager.API.TokenApp = "quicd"
	cmc.CheckManager.Broker.ID = "2"

	// Circonus API Token key (https://login.circonus.com/user/tokens)
	if circonusAPIKey == "" {
		cmc.CheckManager.API.TokenKey = os.Getenv("CIRCONUS_API_TOKEN")
	} else {
		cmc.CheckManager.API.TokenKey = circonusAPIKey
	}

	log.Println("Creating new cgm instance")

	metrics, err = cgm.NewCirconusMetrics(cmc)
	if err != nil {
		logger.Println(err)
		return errors.Wrap(err, "unable to initialize metrics")
	}

	logger.Println("Adding ctrl-c trap")
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		logger.Println("Received CTRL-C, flushing outstanding metrics before exit")
		metrics.Flush()
		os.Exit(0)
	}()

	logger.Println("Starting to send metrics")

	return nil
}
