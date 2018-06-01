package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

func faasPost() func(http.ResponseWriter, *http.Request) {
	baseURL, _ := url.Parse(os.Getenv("base_url"))
	proxy := NewHTTPClientReverseProxy(baseURL, time.Second*30)
	functionSuffix := os.Getenv("function_suffix")

	functionAs := FunctionAsHostBaseURLResolver{
		FunctionSuffix: functionSuffix,
	}

	return MakeForwardingProxyHandler(proxy,
		[]HTTPNotifier{LoggingNotifier{}},
		functionAs)
}

// NewHTTPClientReverseProxy proxies to an upstream host through the use of a http.Client
func NewHTTPClientReverseProxy(baseURL *url.URL, timeout time.Duration) *HTTPClientReverseProxy {
	h := HTTPClientReverseProxy{
		BaseURL: baseURL,
		Timeout: timeout,
	}

	h.Client = &http.Client{}

	return &h
}

// HTTPClientReverseProxy proxy to a remote BaseURL using a http.Client
type HTTPClientReverseProxy struct {
	BaseURL *url.URL
	Client  *http.Client
	Timeout time.Duration
}

// HTTPNotifier notify about HTTP request/response
type HTTPNotifier interface {
	Notify(method string, URL string, statusCode int, duration time.Duration)
}

// BaseURLResolver URL resolver for upstream requests
type BaseURLResolver interface {
	Resolve(r *http.Request) string
}

// MakeForwardingProxyHandler create a handler which forwards HTTP requests
func MakeForwardingProxyHandler(proxy *HTTPClientReverseProxy, notifiers []HTTPNotifier, baseURLResolver BaseURLResolver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		baseURL := baseURLResolver.Resolve(r)

		requestURL := r.URL.Path

		start := time.Now()

		statusCode, err := forwardRequest(w, r, proxy.Client, baseURL, requestURL, proxy.Timeout)

		seconds := time.Since(start)
		if err != nil {
			log.Printf("error with upstream request to: %s, %s\n", requestURL, err.Error())
		}
		for _, notifier := range notifiers {
			notifier.Notify(r.Method, requestURL, statusCode, seconds)
		}
	}
}

func buildUpstreamRequest(r *http.Request, url string) *http.Request {

	if len(r.URL.RawQuery) > 0 {
		url = fmt.Sprintf("%s?%s", url, r.URL.RawQuery)
	}

	upstreamReq, _ := http.NewRequest(r.Method, url, nil)
	copyHeaders(upstreamReq.Header, &r.Header)

	upstreamReq.Header["X-Forwarded-For"] = []string{r.RemoteAddr}

	if r.Body != nil {
		upstreamReq.Body = r.Body
	}

	return upstreamReq
}

func forwardRequest(w http.ResponseWriter, r *http.Request, proxyClient *http.Client, baseURL string, requestURL string, timeout time.Duration) (int, error) {

	upstreamReq := buildUpstreamRequest(r, baseURL+requestURL)
	if upstreamReq.Body != nil {
		defer upstreamReq.Body.Close()
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	res, resErr := proxyClient.Do(upstreamReq.WithContext(ctx))
	if resErr != nil {
		badStatus := http.StatusBadGateway
		w.WriteHeader(badStatus)
		return badStatus, resErr
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	copyHeaders(w.Header(), &res.Header)

	// Write status code
	w.WriteHeader(res.StatusCode)

	if res.Body != nil {
		// Copy the body over
		io.CopyBuffer(w, res.Body, nil)
	}

	return res.StatusCode, nil
}

func copyHeaders(destination http.Header, source *http.Header) {
	for k, v := range *source {
		vClone := make([]string, len(v))
		copy(vClone, v)
		(destination)[k] = vClone
	}
}

func getServiceName(urlValue string) string {
	var serviceName string
	forward := "/function/"
	if strings.HasPrefix(urlValue, forward) {
		serviceName = urlValue[len(forward):]
	}
	return serviceName
}

// LoggingNotifier notifies a log about a request
type LoggingNotifier struct {
}

// Notify a log about a request
func (LoggingNotifier) Notify(method string, URL string, statusCode int, duration time.Duration) {
	log.Printf("Forwarded [%s] to %s - [%d] - %f seconds", method, URL, statusCode, duration.Seconds())
}

// FunctionAsHostBaseURLResolver resolves URLs using a function from the URL as a host
type FunctionAsHostBaseURLResolver struct {
	FunctionSuffix string
}

// Resolve the base URL for a request
func (f FunctionAsHostBaseURLResolver) Resolve(r *http.Request) string {

	query := r.URL.Query()
	fn := query.Get("fn")

	return fmt.Sprintf("http://%s", fn)
}
