package main

import (
	"fmt"
	"net/http"
	"errors"
)

type Response struct {
	statusCode int
	body       string
	headers    http.Header
}

func (r *Response) Write(w http.ResponseWriter) {
	head := w.Header()
	for k, _ := range r.headers {
		head.Set(k, r.headers.Get(k))
	}
	w.WriteHeader(r.statusCode)
	fmt.Fprint(w, r.body)
}

func ResponseFromError(err error) *Response {
	return &Response{
		statusCode: 500,
		body:       err.Error(),
	}
}

type RequestHandler func(*http.Request) *Response
type ResponseHandler func(*Response) error

type Handler struct {
	Requests  []RequestHandler
	Responses []ResponseHandler
	Handler   RequestHandler
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) AddMiddleware(mid Middleware) error {
	h.Requests = append(h.Requests, mid.GetRequest())
	h.Responses = append(h.Responses, mid.GetResponse())
	return nil
}

func (h *Handler) SetHandler(hand RequestHandler) error {
	if h.Handler != nil {
		return errors.New("Setting handler twice!!")
	}
	h.Handler = hand
	return nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	for _, reqHandler := range h.Requests {
		resp := reqHandler(req)
		if resp != nil {
			resp.Write(w)
			return
		}
	}
	resp := h.Handler(req)
	for _, respHandler := range h.Responses {
		err := respHandler(resp)
		if err != nil {
			ResponseFromError(err).Write(w)
			return
		}
	}
	resp.Write(w)
}

type Middleware interface {
	GetRequest() RequestHandler
	GetResponse() ResponseHandler
}

type BasicMiddleware struct {
	req  RequestHandler
	resp ResponseHandler
}

func (mid * BasicMiddleware) GetRequest() RequestHandler {
	return mid.req
}

func (mid * BasicMiddleware) GetResponse() ResponseHandler {
	return mid.resp
}

func NewBasicMiddleware(req RequestHandler, resp ResponseHandler) Middleware {
	return &BasicMiddleware{
		req:  req,
		resp: resp,
	}
}
