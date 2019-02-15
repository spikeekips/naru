package rest

import (
	goLog "log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	logging "github.com/inconshreveable/log15"
	"github.com/spikeekips/naru/cache"
	"github.com/spikeekips/naru/sebak"
	"github.com/spikeekips/naru/storage"
	"golang.org/x/net/http2"

	sebakcommon "boscoin.io/sebak/lib/common"
	sebaknode "boscoin.io/sebak/lib/node"

	cachebackend "github.com/spikeekips/naru/cache/backend"
)

type Server struct {
	bind      *sebakcommon.Endpoint
	st        *storage.Storage
	sst       *sebak.Storage
	cch       *cache.Cache
	sebakInfo sebaknode.NodeInfo
	core      *http.Server
	log       logging.Logger
	router    *mux.Router
}

func NewServer(bind *sebakcommon.Endpoint, st *storage.Storage, sst *sebak.Storage, cb cachebackend.Backend, sebakInfo sebaknode.NodeInfo) *Server {
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

	cch := cache.NewCache("api", cb)

	server := &Server{
		bind:      bind,
		st:        st,
		sst:       sst,
		cch:       cch,
		sebakInfo: sebakInfo,
		core:      core,
		log:       httpLog,
		router:    mux.NewRouter(),
	}

	server.router.Use(FlushWriterMiddleware())
	core.Handler = HTTP2Log15Handler{log: httpLog, handler: server.router}

	server.addDefaultHandlers()

	return server
}

func (s *Server) AddHandler(pattern string, handler func(http.ResponseWriter, *http.Request)) *mux.Route {
	return s.router.HandleFunc(pattern, handler)
}

func (s *Server) addDefaultHandlers() {
	restHandler := NewHandler(s.st, s.sst, s.cch, s.sebakInfo)

	s.AddHandler("/", restHandler.Index)
	s.AddHandler("/api/v1/accounts", restHandler.GetAccounts).
		Methods("POST").
		Headers("Content-Type", "application/json")
	s.AddHandler(
		"/api/v1/accounts/{id}",
		NewCacheHandler(s.cch, time.Second*3, restHandler.GetAccount).
			Status(time.Second*3, http.StatusNotFound). // no-cache
			Status(0, http.StatusOK).                   // for 1 year
			Status(-1).                                 // no-cache for other status
			SetCacheKey(func(r *http.Request) string {
				return r.URL.Path
			}).
			Handler(),
	).
		Methods("GET")
	s.AddHandler(
		"/api/v1/blocks/{hashOrHeight}",
		NewCacheHandler(s.cch, time.Second*3, restHandler.GetBlock).
			Status(time.Second*3, http.StatusNotFound). // no-cache
			Status(0, http.StatusOK).                   // permanent
			Status(-1).                                 // no-cache for other status
			SetCacheKey(func(r *http.Request) string {
				return r.URL.Path
			}).
			Handler(),
	).
		Methods("GET")
	s.AddHandler("/api/v1/transactions", restHandler.PostTransaction).
		Methods("POST").
		Headers("Content-Type", "application/json")
	s.AddHandler(
		"/api/v1/transactions/{id}/status",
		NewCacheHandler(s.cch, time.Second*3, restHandler.GetTransactionStatus).
			Status(0, http.StatusOK). // permanent
			Status(time.Second*3, http.StatusAccepted).
			Status(time.Second*5, http.StatusNotFound). // no-cache
			Status(-1).                                 // no-cache for other status
			SetCacheKey(func(r *http.Request) string {
				return r.URL.Path
			}).
			Handler(),
	).
		Methods("Get")
	s.AddHandler(
		"/api/v1/transactions/{id}",
		NewCacheHandler(s.cch, time.Second*3, restHandler.GetTransactionByHash).
			Status(time.Second*3, http.StatusNotFound). // no-cache
			Status(0, http.StatusOK).                   // for 1 year
			Status(-1).                                 // no-cache for other status
			SetCacheKey(func(r *http.Request) string {
				return r.URL.Path
			}).
			Handler(),
	).Methods("Get")
	s.AddHandler(
		"/api/v1/accounts/{id}/operations",
		NewStreamer(OperationsByAccountStreamHandler{H: restHandler}, time.Second*10).Handler,
	).
		Headers("Accept", "text/event-stream").
		Methods("GET")
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

	err := listenFunc()
	if err != nil {
		if err == http.ErrServerClosed {
			return nil
		}
	}

	return err
}

func (s *Server) Stop() error {
	return s.core.Close()
}
