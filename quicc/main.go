package main

import (
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	stdlog "log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/lucas-clemente/quic-go/h2quic"
	isatty "github.com/mattn/go-isatty"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	logFormatFlag := flag.String("log-format", "auto", "log format")
	logLevelFlag := flag.String("log-level", "DEBUG", "log level")

	flag.Parse()
	urls := flag.Args()
	if len(urls) == 0 {
		fmt.Println("Need one or more URLs to download")
		os.Exit(1)
	}

	initLog(*logFormatFlag, *logLevelFlag)

	quicTransport := &h2quic.QuicRoundTripper{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	hclient := &http.Client{
		Transport: quicTransport,
	}

	var wg sync.WaitGroup
	wg.Add(len(urls))
	for _, addr := range urls {
		log.Info().Str("addr", addr).Msg("GET request")
		go func(addr string) {
			rsp, err := hclient.Get(addr)
			if err != nil {
				panic(err)
			}

			body := &bytes.Buffer{}
			_, err = io.Copy(body, rsp.Body)
			if err != nil {
				panic(err)
			}
			log.Info().Str("addr", addr).Str("rsp", rsp.Status).Int("body-bytes", len(body.Bytes())).Msg("GET Response")
			wg.Done()
		}(addr)
	}
	wg.Wait()
}

func initLog(format, level string) {
	switch l := strings.ToLower(level); l {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	default:
		panic(fmt.Sprintf("unsupported log level: %q", l))
	}

	// os.Stdout isn't guaranteed to be thread-safe, wrap in a sync writer.
	// Files are guaranteed to be safe, terminals are not.
	var logWriter io.Writer
	if isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()) {
		logWriter = zerolog.SyncWriter(os.Stdout)
	} else {
		logWriter = os.Stdout
	}

	if format == "auto" {
		if isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()) {
			format = "human"
		} else {
			format = "json"
		}
	}

	var zlog zerolog.Logger
	switch format {
	case "json":
		zlog = zerolog.New(logWriter).With().Timestamp().Logger()
	case "human":
		useColor := true
		w := zerolog.ConsoleWriter{
			Out:     logWriter,
			NoColor: !useColor,
		}
		zlog = zerolog.New(w).With().Timestamp().Logger()
	default:
		log.Error().Str("format", format).Msg("unsupported log format")
		os.Exit(1)
	}

	log.Logger = zlog

	stdlog.SetFlags(0)
	stdlog.SetOutput(zlog)
}
