// Package httpserver ...
package httpserver

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/n-r-w/log-server-v2/pkg/logger"
)

const (
	defaultReadTimeout     = 5 * time.Second
	defaultWriteTimeout    = 5 * time.Second
	defaultAddr            = ":8080"
	defaultShutdownTimeout = 3 * time.Second
)

type Server struct {
	server          *http.Server
	logger          logger.Interface
	notify          chan error
	shutdownTimeout time.Duration
}

func New(handler http.Handler, logger logger.Interface, opts ...Option) *Server {
	httpServer := &http.Server{
		Handler:      handler,
		ReadTimeout:  defaultReadTimeout,
		WriteTimeout: defaultWriteTimeout,
		Addr:         defaultAddr,
	}

	s := &Server{
		server:          httpServer,
		logger:          logger,
		notify:          make(chan error, 1),
		shutdownTimeout: defaultShutdownTimeout,
	}

	for _, opt := range opts {
		opt(s)
	}

	s.start()

	return s
}

func (s *Server) start() {
	go func() {
		l, err := net.Listen("tcp", s.server.Addr)
		if err == nil {
			s.logger.Info("server started on port %s", s.server.Addr)
			err = s.server.Serve(l)
		}
		s.notify <- err
		close(s.notify)
	}()
}

func (s *Server) Notify() <-chan error {
	return s.notify
}

func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()

	return s.server.Shutdown(ctx)
}
