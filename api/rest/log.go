package rest

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	logging "github.com/inconshreveable/log15"

	sebakcommon "boscoin.io/sebak/lib/common"
	sebakmetrics "boscoin.io/sebak/lib/metrics"
)

type HTTP2ErrorLog15Writer struct {
	l logging.Logger
}

func (w HTTP2ErrorLog15Writer) Write(b []byte) (int, error) {
	w.l.Error("error", "error", string(b))
	return 0, nil
}

type HTTP2ResponseLog15Writer struct {
	http.ResponseWriter
	status        int
	size          int
	closeNotifier http.CloseNotifier
	flusher       http.Flusher
}

func NewHTTP2ResponseLog15Writer(w http.ResponseWriter) *HTTP2ResponseLog15Writer {
	closeNotifier, ok := w.(http.CloseNotifier)
	if !ok {
		closeNotifier = FakeCloseNotifier{}
		log.Warn("FakeCloseNotifier used")
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		flusher = FakeFlusher{}
		log.Warn("FakeFlusher used")
	}

	return &HTTP2ResponseLog15Writer{
		ResponseWriter: w,
		closeNotifier:  closeNotifier,
		flusher:        flusher,
	}
}

func (l *HTTP2ResponseLog15Writer) Write(b []byte) (int, error) {
	size, err := l.ResponseWriter.Write(b)
	l.size += size
	return size, err
}

func (l *HTTP2ResponseLog15Writer) WriteHeader(s int) {
	l.ResponseWriter.WriteHeader(s)
	l.status = s
}

func (l *HTTP2ResponseLog15Writer) Status() int {
	if l.status == 0 {
		return 200
	}
	// when it doesn't call WriteHeader, default status is 200.
	return l.status
}

func (l *HTTP2ResponseLog15Writer) Size() int {
	return l.size
}

func (l *HTTP2ResponseLog15Writer) Flush() {
	l.flusher.Flush()
}

func (l *HTTP2ResponseLog15Writer) CloseNotify() <-chan bool {
	return l.closeNotifier.CloseNotify()
}

type HTTP2Log15Handler struct {
	log     logging.Logger
	handler http.Handler
}

var HeaderKeyFiltered []string = []string{
	"Content-Length",
	"Content-Type",
	"Accept",
	"Accept-Encoding",
	"User-Agent",
}

// ServeHTTP will log in 2 phase, when request received and response sent. This
// was derived from github.com/gorilla/handlers/handlers.go
func (l HTTP2Log15Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	begin := time.Now()
	uid := sebakcommon.GenerateUUID()

	uri := r.RequestURI
	if r.ProtoMajor == 2 && r.Method == "CONNECT" {
		uri = r.Host
	}
	if uri == "" {
		uri = r.URL.RequestURI()
	}

	header := http.Header{}
	for key, value := range r.Header {
		if _, found := sebakcommon.InStringArray(HeaderKeyFiltered, key); found {
			continue
		}
		header[key] = value
	}

	l.log.Debug(
		"request",
		"content-length", r.ContentLength,
		"content-type", r.Header.Get("Content-Type"),
		"headers", header,
		"host", r.Host,
		"id", uid,
		"method", r.Method,
		"proto", r.Proto,
		"referer", r.Referer(),
		"remote", r.RemoteAddr,
		"uri", uri,
		"user-agent", r.UserAgent(),
		"x-forwarded-for", strings.Split(r.Header.Get("X-Forwarded-For"), ",")[0],
	)

	writer := NewHTTP2ResponseLog15Writer(w)
	l.handler.ServeHTTP(writer, r)

	elapsed := time.Since(begin)

	l.log.Debug(
		"response",
		"id", uid,
		"status", writer.Status(),
		"size", writer.Size(),
		"elapsed", elapsed,
	)

	{
		labels := []string{"path", r.URL.Path, "method", r.Method, "status", strconv.Itoa(writer.Status())}
		sebakmetrics.API.RequestDurationSeconds.With(labels...).Observe(elapsed.Seconds())
		sebakmetrics.API.RequestsTotal.With(labels...).Add(1)

		if writer.Status() >= 500 {
			sebakmetrics.API.RequestErrorsTotal.With(labels...).Add(1)
		}
	}
}
