package restv1

import (
	"fmt"
	"net/http"
	"time"
)

type StreamHandler interface {
	Init() <-chan interface{}
	Stream() (<-chan interface{}, func())
}

type BaseStreamHandler struct {
	w http.ResponseWriter
	r *http.Request
}

func (b BaseStreamHandler) Request() *http.Request {
	return b.r
}

func (b BaseStreamHandler) ResponseWriter() http.ResponseWriter {
	return b.w
}

type NewStreamHandler interface {
	NewRequest(BaseStreamHandler) (StreamHandler, error)
}

type Streamer struct {
	newHandler NewStreamHandler
	timeout    time.Duration
}

func NewStreamer(newHandler NewStreamHandler, timeout time.Duration) Streamer {
	return Streamer{newHandler: newHandler, timeout: timeout}
}

func (s Streamer) Handler(w http.ResponseWriter, r *http.Request) {
	var connCloseNotify <-chan bool
	if cw, ok := w.(http.CloseNotifier); !ok {
		panic(fmt.Errorf("http.CloseNotifier not found"))
	} else {
		connCloseNotify = cw.CloseNotify()
	}

	jw := NewJSONWriter(w)
	jw.Header().Set("X-SEBAK-TIMEOUT", s.timeout.String())

	streamer, err := s.newHandler.NewRequest(
		BaseStreamHandler{w: jw, r: r},
	)
	if err != nil {
		jw.WriteObject(err)
		return
	}

	initChan := streamer.Init()
	streamChan, closeStreamFunc := streamer.Stream()
	timeoutChan := time.After(s.timeout)
	streamReadyChan := make(chan bool)
	stopChan := make(chan bool)
	defer close(streamReadyChan)
	defer close(stopChan)
	defer closeStreamFunc()

	go func() {
	initEnd:
		for {
			select {
			case v, ok := <-initChan:
				if !ok {
					break initEnd
				}
				switch v.(type) {
				case error:
					jw.WriteObject(v)
					stopChan <- true
					break initEnd
				case bool:
					streamReadyChan <- true
				default:
					jw.WriteObject(v)
				}
			}
		}
	}()

	var streamBuffer []interface{}
	var streamReady bool
streamEnd:
	for {
		select {
		case <-timeoutChan:
			log.Debug("timeout", "timeout", s.timeout)
			break streamEnd
		case <-connCloseNotify:
			log.Debug("HTTP connection just closed from client-side")
			break streamEnd
		case <-stopChan:
			break streamEnd
		case <-streamReadyChan:
			if len(streamBuffer) > 0 {
				for _, i := range streamBuffer {
					jw.WriteObject(i)
				}
				streamBuffer = nil
			}
			streamReady = true
		case v, ok := <-streamChan:
			if !ok {
				break streamEnd
			}

			switch v.(type) {
			case error:
				jw.WriteObject(v)
				break streamEnd
			}

			if !streamReady {
				streamBuffer = append(streamBuffer, v)
				continue
			}

			jw.WriteObject(v)
			if _, ok := v.(error); ok {
				break streamEnd
			}
		}
	}
}
