package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"strconv"

	"github.com/bcap/flamegrapher/assets"
)

const DataPath = "data.json"

type Server struct {
	*http.Server
}

func NewServer(port int, compressedJSONData []byte) Server {
	return Server{
		Server: &http.Server{
			Addr:    ":" + strconv.Itoa(port),
			Handler: handler(compressedJSONData),
		},
	}
}

func handler(compressedJSONData []byte) http.HandlerFunc {
	fs := http.FS(assets.FS)
	staticHandler := http.FileServer(fs)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/"+DataPath {
			staticHandler.ServeHTTP(w, r)
			return
		}

		w.Header().Add("Content-type", "application/json")
		w.Header().Add("Content-encoding", "gzip")
		w.WriteHeader(http.StatusOK)
		w.Write(compressedJSONData)
	})
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
	port := listener.Addr().(*net.TCPAddr).Port
	log.Printf("Access the flamegraph at http://localhost:%d", port)
	return s.Serve(listener)
}
