package rest

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/prometheus/common/log"

	sebakerrors "boscoin.io/sebak/lib/errors"
	sebakhttputils "boscoin.io/sebak/lib/network/httputils"
)

type FakeCloseNotifier struct {
}

func (f FakeCloseNotifier) CloseNotify() <-chan bool {
	return nil
}

type FakeFlusher struct {
}

func (f FakeFlusher) Flush() {
}

type FlushWriter struct {
	http.ResponseWriter
	flusher http.Flusher
}

func NewFlushWriter(w http.ResponseWriter) FlushWriter {
	var flusher http.Flusher = FakeFlusher{}
	if f, ok := w.(http.Flusher); ok {
		flusher = f
	} else {
		log.Warn("FakeFlusher used")
	}

	return FlushWriter{ResponseWriter: w, flusher: flusher}
}

func (fw FlushWriter) Write(p []byte) (n int, err error) {
	n, err = fw.ResponseWriter.Write(p)
	fw.flusher.Flush()

	return
}

func (fw FlushWriter) Flush() {
	fw.flusher.Flush()
}

func (fw FlushWriter) CloseNotify() <-chan bool {
	f, ok := fw.ResponseWriter.(http.CloseNotifier)
	if !ok {
		return nil
	}

	return f.CloseNotify()
}

func FlushWriterMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, ok := w.(FlushWriter); !ok {
				w = NewFlushWriter(w)
			}
			next.ServeHTTP(w, r)
		})
	}
}

type JSONWriter struct {
	http.ResponseWriter
}

func NewJSONWriter(w http.ResponseWriter) *JSONWriter {
	return &JSONWriter{ResponseWriter: w}
}

func (j *JSONWriter) writeObject(v interface{}) (int, error) {
	bs, err := json.Marshal(v)
	if err != nil {
		return 0, err
	}

	return j.Write(append(bs, []byte("\n")...))
}

var ErrorsToStatus = map[uint]int{
	sebakerrors.StorageRecordDoesNotExist.Code: http.StatusNotFound,
	sebakerrors.TransactionNotFound.Code:       http.StatusNotFound,
}

func (j *JSONWriter) statusByError(err error) int {
	var se *sebakerrors.Error
	var ok bool
	if se, ok = err.(*sebakerrors.Error); ok {
		if c, ok := ErrorsToStatus[se.Code]; ok {
			return c
		}

		if c := se.GetData("status"); c != nil {
			return c.(int)
		}
	}

	return sebakhttputils.StatusCode(err)
}

func (j *JSONWriter) WriteObject(v interface{}) (int, error) {
	if h, ok := v.(sebakhttputils.HALResource); ok {
		j.Header().Set("Content-Type", "application/hal+json")
		v = h.Resource()
	} else if e, ok := v.(error); ok {
		j.Header().Set("Content-Type", "application/problem+json")

		status := j.statusByError(e)
		j.WriteHeader(status)
		v = sebakhttputils.NewErrorProblem(e, status)
	} else {
		j.Header().Set("Content-Type", "application/json")
	}

	return j.writeObject(v)
}
