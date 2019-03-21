package common

import (
	"bytes"
	"context"
	"crypto/tls"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/http2"
)

type HTTP2Client struct {
	timeout   time.Duration
	url       *url.URL
	client    http.Client
	transport *http.Transport
	headers   http.Header
	keepAlive bool
}

func NewHTTP2Client(timeout time.Duration, url *url.URL, keepAlive bool, headers http.Header) (http2Client *HTTP2Client, err error) {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		IdleConnTimeout:   timeout,
		DisableKeepAlives: !keepAlive,
		DialContext: (&net.Dialer{
			Timeout:   timeout,
			DualStack: true,
		}).DialContext,
	}

	if err = http2.ConfigureTransport(transport); err != nil {
		return
	}

	client := http.Client{
		Transport: transport,
		Timeout:   timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // NOTE prevent redirect
		},
	}
	http2Client = &HTTP2Client{
		timeout:   timeout,
		url:       url,
		client:    client,
		transport: transport,
		headers:   headers,
		keepAlive: keepAlive,
	}

	return
}

func (cl *HTTP2Client) URL() *url.URL {
	return cl.url
}

func (cl *HTTP2Client) Transport() *http.Transport {
	return cl.transport
}

func (cl *HTTP2Client) resolvePath(path string) *url.URL {
	return cl.url.ResolveReference(&url.URL{Path: path})
}

func (cl *HTTP2Client) newHeaders(headers http.Header) http.Header {
	newHeaders := http.Header{}
	for k, v := range cl.headers {
		newHeaders[k] = v
	}

	if headers != nil {
		for k, v := range headers {
			newHeaders[k] = v
		}
	}

	return newHeaders
}

func (cl *HTTP2Client) request(method, path string, body io.Reader, headers http.Header) (response *http.Response, err error) {
	u := cl.resolvePath(path)

	var r *http.Request
	if r, err = http.NewRequest(method, u.String(), body); err != nil {
		return
	}
	defer func() {
		r.Close = true
	}()

	r.Header = cl.newHeaders(headers)

	if cl.timeout > 0 {
		ctx, _ := context.WithTimeout(context.TODO(), cl.timeout)
		r = r.WithContext(ctx)
	}

	response, err = cl.client.Do(r)

	return
}

func (cl *HTTP2Client) Get(path string, headers http.Header) (b []byte, err error) {
	var response *http.Response
	if response, err = cl.request("GET", path, nil, headers); err != nil {
		return
	}
	defer response.Body.Close()

	b, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}

	if response.StatusCode != http.StatusOK {
		err = HTTPProblem.New().SetData("status", response.StatusCode).SetData("body", string(b))
		return
	}

	return
}

func (cl *HTTP2Client) Post(path string, body []byte, headers http.Header) (b []byte, err error) {
	var bodyReader io.Reader

	if body != nil {
		bodyReader = bytes.NewBuffer(body)
	}

	var response *http.Response
	if response, err = cl.request("POST", path, bodyReader, headers); err != nil {
		return
	}
	defer response.Body.Close()

	b, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}

	if response.StatusCode != http.StatusOK {
		err = HTTPProblem.New().SetData("status", response.StatusCode).SetData("body", string(b))
		return
	}

	return
}
