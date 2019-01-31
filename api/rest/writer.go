package rest

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	sebakhttputils "boscoin.io/sebak/lib/network/httputils"
)

type FlushWriter struct {
	w http.ResponseWriter
}

func (fw FlushWriter) Header() http.Header {
	return fw.w.Header()
}

func (fw FlushWriter) WriteHeader(status int) {
	fw.w.WriteHeader(status)
}

func (fw FlushWriter) Write(p []byte) (n int, err error) {
	n, err = fw.w.Write(p)
	if f, ok := fw.w.(http.Flusher); ok {
		f.Flush()
	}
	return
}

func FlushWriterMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, ok := w.(FlushWriter); !ok {
				w = FlushWriter{w}
			}
			next.ServeHTTP(w, r)
		})
	}
}

type JSONWriter struct {
	status     int
	w          http.ResponseWriter
	sentHeader bool
}

func (j *JSONWriter) setHeader(name, value string) {
	if j.sentHeader {
		return
	}
	j.w.Header().Set(name, value)
}

func (j *JSONWriter) Write(v interface{}) error {
	if h, ok := v.(sebakhttputils.HALResource); ok {
		j.setHeader("Content-Type", "application/hal+json")
		v = h.Resource()
	} else if e, ok := v.(error); ok {
		j.setHeader("Content-Type", "application/problem+json")
		v = sebakhttputils.NewErrorProblem(e, j.status)
	} else {
		j.setHeader("Content-Type", "application/json")
	}

	if !j.sentHeader {
		j.w.WriteHeader(j.status)
	}

	j.sentHeader = true

	bs, err := json.Marshal(v)
	if err != nil {
		return err
	}

	if _, err := j.w.Write(bs); err != nil {
		return err
	}
	fmt.Fprintln(j.w)

	return nil
}
