package adapter

import "net/http"

// define custom response
type ResponseRecorder struct {
	Status int
	Body   []byte
}

func (r *ResponseRecorder) Header() http.Header {
	return http.Header{}
}

func (r *ResponseRecorder) Write(data []byte) (int, error) {
	r.Body = append(r.Body, data...)
	return len(data), nil
}

func (r *ResponseRecorder) WriteHeader(statusCode int) {
	r.Status = statusCode
}
