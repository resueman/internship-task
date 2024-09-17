package http_server

import (
	"context"
	"net/http"
	"time"
)

type Server struct {
	server *http.Server
	notify chan error
}

func New(handler http.Handler, address string) *Server {
	httpServer := &http.Server{
		Handler: handler,
		Addr:    address,
	}

	s := &Server{
		server: httpServer,
		notify: make(chan error, 1),
	}

	s.start()

	return s
}

func (s *Server) start() {
	go func() {
		s.notify <- s.server.ListenAndServe()
		close(s.notify)
	}()
}

func (s *Server) Notify() <-chan error {
	return s.notify
}

func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return s.server.Shutdown(ctx)
}
