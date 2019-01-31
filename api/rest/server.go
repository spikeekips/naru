package rest

import (
	goLog "log"
	"net/http"

	"github.com/gorilla/mux"
	logging "github.com/inconshreveable/log15"
	"golang.org/x/net/http2"

	sebakcommon "boscoin.io/sebak/lib/common"
)

type Server struct {
	bind   *sebakcommon.Endpoint
	core   *http.Server
	log    logging.Logger
	router *mux.Router
}

func NewServer(bind *sebakcommon.Endpoint) *Server {
	core := &http.Server{
		Addr: bind.Host,
		/* TODO from config
		ReadTimeout:       config.ReadTimeout,
		ReadHeaderTimeout: config.ReadHeaderTimeout,
		WriteTimeout:      config.WriteTimeout,
		*/
		ErrorLog: goLog.New(HTTP2ErrorLog15Writer{httpLog}, "", 0),
	}
	core.SetKeepAlivesEnabled(true)

	http2.ConfigureServer(
		core,
		&http2.Server{
			// MaxConcurrentStreams
			// MaxReadFrameSize
			// IdleTimeout
		},
	)

	server := &Server{
		bind:   bind,
		core:   core,
		log:    httpLog,
		router: mux.NewRouter(),
	}

	core.Handler = HTTP2Log15Handler{log: httpLog, handler: server.router}

	return server
}

func (s *Server) Start() error {
	var listenFunc func() error
	switch s.bind.Scheme {
	case "http":
		listenFunc = func() error {
			return s.core.ListenAndServe()
		}
	case "https":
		listenFunc = func() error {
			return s.core.ListenAndServeTLS(
				s.bind.Query().Get("tls-cert"),
				s.bind.Query().Get("tls-key"),
			)
		}
	}

	return listenFunc()
}

func (s *Server) AddHandler(pattern string, handler func(http.ResponseWriter, *http.Request)) *mux.Route {
	return s.router.HandleFunc(pattern, handler)
}
