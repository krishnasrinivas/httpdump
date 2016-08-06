package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/vulcand/oxy/forward"
)

type responseWriterTrap struct {
	writer http.ResponseWriter
}

func (w responseWriterTrap) Header() http.Header {
	return w.writer.Header()
}

func (w responseWriterTrap) Write(p []byte) (int, error) {
	os.Stdout.Write(p)
	return w.writer.Write(p)
}

func (w responseWriterTrap) WriteHeader(i int) {
	fmt.Printf("\n-----\n")
	fmt.Printf("RESPONSE STATUS: %d %s\n", i, http.StatusText(i))
	for k, v := range w.writer.Header() {
		fmt.Printf("%s: %s\n", k, v[0])
	}
	fmt.Println()
	w.writer.WriteHeader(i)
}

type logger struct {
	h http.Handler
}

func (l logger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("\n---------------------------\n")
	fmt.Printf("REQUEST : %s %s\n", r.Method, r.URL)
	fmt.Printf("Host: %s\n", r.Host)
	for k, v := range r.Header {
		fmt.Printf("%s: %s\n", k, v[0])
	}
	fmt.Println()
	l.h.ServeHTTP(responseWriterTrap{w}, r)
	fmt.Printf("\n--------------------------\n")
}

type forwarder struct {
	scheme string
	host   string
	h      http.Handler
}

type closer func() error

func (c closer) Close() error {
	return c()
}

func (f forwarder) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.URL.Scheme = f.scheme
	r.URL.Host = f.host
	body := r.Body
	r.Body = struct {
		io.Reader
		io.Closer
	}{
		io.TeeReader(body, os.Stdout),
		closer(func() error {
			return body.Close()
		}),
	}
	f.h.ServeHTTP(w, r)
}

func main() {
	listenAddr := flag.String("l", ":8000", "listen address")
	fwdAddr := flag.String("f", "localhost:9000", "forward address")
	flag.Parse()
	fwd, _ := forward.New(forward.PassHostHeader(true))
	server := &http.Server{
		Addr:    *listenAddr,
		Handler: logger{forwarder{"http", *fwdAddr, fwd}},
	}
	server.ListenAndServe()
}
