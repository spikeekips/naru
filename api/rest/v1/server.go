package restv1

import (
	goLog "log"
	"net/http"
	"time"

	sebakcommon "boscoin.io/sebak/lib/common"
	sebaknode "boscoin.io/sebak/lib/node"
	"github.com/gorilla/mux"
	logging "github.com/inconshreveable/log15"
	"golang.org/x/net/http2"

	"github.com/spikeekips/naru/api/rest"
	"github.com/spikeekips/naru/cache"
	cachebackend "github.com/spikeekips/naru/cache/backend"
	"github.com/spikeekips/naru/config"
	"github.com/spikeekips/naru/element"
	"github.com/spikeekips/naru/sebak"
)

type Server struct {
	bind      *sebakcommon.Endpoint
	sst       *sebak.Storage
	potion    element.Potion
	cch       *cache.Cache
	sebakInfo sebaknode.NodeInfo
	core      *http.Server
	log       logging.Logger
	router    *mux.Router
}

func NewServer(nc *config.Network, sst *sebak.Storage, potion element.Potion, cb cachebackend.Backend, sebakInfo sebaknode.NodeInfo) *Server {
	httpLog := logging.New("module", "restv1")
	nc.Log.HTTP.SetLogger(httpLog)

	core := &http.Server{
		Addr: nc.Bind.Host,
		/* TODO from config
		ReadTimeout:       config.ReadTimeout,
		ReadHeaderTimeout: config.ReadHeaderTimeout,
		WriteTimeout:      config.WriteTimeout,
		*/
		ErrorLog: goLog.New(rest.HTTP2ErrorLog15Writer{L: httpLog}, "", 0),
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
		bind:      nc.Bind,
		sst:       sst,
		potion:    potion,
		cch:       cch,
		sebakInfo: sebakInfo,
		core:      core,
		log:       httpLog,
		router:    mux.NewRouter(),
	}

	// TODO ratelimit
	server.router.Use(rest.FlushWriterMiddleware())
	server.router.Use(rest.PotionMiddleware(potion))
	core.Handler = rest.HTTP2Log15Handler{Log: httpLog, Handler: server.router}

	server.addDefaultHandlers()

	return server
}

func (s *Server) AddHandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) *mux.Route {
	return s.router.HandleFunc(pattern, handler)
}

func (s *Server) AddHandler(pattern string, handler http.Handler) *mux.Route {
	return s.router.Handle(pattern, handler)
}

func (s *Server) addDefaultHandlers() {
	restHandler := NewHandler(s.sst, s.potion, s.cch, s.sebakInfo)

	s.AddHandleFunc("/", restHandler.Index)
	s.AddHandleFunc("/api/v1/accounts", restHandler.GetAccounts).
		Methods("POST").
		Headers("Content-Type", "application/json")
	s.AddHandleFunc(
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
	s.AddHandleFunc(
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
	s.AddHandleFunc("/api/v1/transactions", restHandler.PostTransaction).
		Methods("POST").
		Headers("Content-Type", "application/json")
	s.AddHandleFunc(
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
	s.AddHandleFunc(
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
	s.AddHandleFunc(
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
