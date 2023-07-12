package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"

	"github.com/bcap/flamegrapher/assets"
)

type DataFetcher = func(context.Context) (json.RawMessage, error)

const DataPath = "data.json"

type Server struct {
	*http.Server
}

func NewServer(port int, fetcher DataFetcher) Server {
	return Server{
		Server: &http.Server{
			Addr:    ":" + strconv.Itoa(port),
			Handler: newHandler(fetcher),
		},
	}
}

func (s Server) Run(ctx context.Context) error {
	listener, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}
	go func() {
		<-ctx.Done()
		_ = s.Close()
	}()
	log.Printf("server listening at http://localhost" + s.Addr)
	return s.Serve(listener)
}

func newHandler(fetcher DataFetcher) http.HandlerFunc {
	fs := http.FS(assets.FS)
	staticHandler := http.FileServer(fs)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// log.Printf("%s %s", r.Method, r.URL)

		if r.URL.Path != "/"+DataPath {
			staticHandler.ServeHTTP(w, r)
			return
		}

		data, err := fetcher(r.Context())
		if err != nil {
			writeErr(w, r, err)
			return
		}

		w.Header().Add("Content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	})
}

func writeErr(w http.ResponseWriter, r *http.Request, err error) {
	w.Header().Add("Content-type", "text/plain")
	w.WriteHeader(http.StatusInternalServerError)
	errBody := fmt.Sprintf("internal error occured: %s", err.Error())
	w.Write([]byte(errBody))
}
