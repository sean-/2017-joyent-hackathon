# Hackathon

A quick hackathon project (3hrs) to demonstrate the value of QUIC.

* `quicc`, a QUIC client.  Connects to a QUIC server, fetches a URL, and
  discards the output.  Useful for benchmarking.
* `quicd`, a QUIC server.  Connects to a backend HTTP server and proxies the
  result via QUIC.  Useful for exposing an existing HTTP(S) service.

## `quicd` Usage

Configure `quicd` to proxy connections to an IP address serving
`www.google.com`:

```
$ cd quicd
$ go build
$ make ssl-keys
$ ./quicd -upstream "http://www.google.com:80" -cert quicd.pem -key quicd.key -circonus-api-key noop
# or
$ ./quicd -upstream "https://www.google.com:443" -cert quicd.pem -key quicd.key -circonus-api-key noop
```

## `quicc` Usage

```
$ cd quicc
$ go build
$ ./quicc https://127.0.0.1:8443/
2017-11-21T08:59:58-08:00 |INFO| GET request addr=https://127.0.0.1:8443/
2017-11-21T08:59:58-08:00 |INFO| GET Response addr=https://127.0.0.1:8443/ rsp="200 OK"
$ ./quicc https://www.google.com:443/
2017-11-21T09:00:19-08:00 |INFO| GET request addr=https://www.google.com:443/
2017-11-21T09:00:19-08:00 |INFO| GET Response addr=https://www.google.com:443/ rsp="200 OK"
$ ./quicc https://www.amazon.com:443/
2017-11-21T09:01:00-08:00 |INFO| GET request addr=https://www.amazon.com:443/
panic: Get https://www.amazon.com:443/: read udp [::]:36103: use of closed network connection

goroutine 5 [running]:
	main.main.func1(0xc42007d0e0, 0xc420016770, 0x7fffffffe7e8, 0x1b) /home/seanc/go/src/github.com/sean-/2017-joyent-hackathon/quicc/main.go:50 +0x1af
reated by main.main
	/home/seanc/go/src/github.com/sean-/2017-joyent-hackathon/quicc/main.go:47 +0x2a5
Exit 2
```
