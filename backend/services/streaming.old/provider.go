package streaming

import (
	"context"
	"errors"
	"io"
	"net/http"
)

var ErrNotFound = errors.New("stream not found")

// Request encapsulates a streaming request coming from the handler layer.
type Request struct {
	Path                string
	RangeHeader         string
	Method              string
	SeekSeconds         float64
	DurationHintSeconds float64
}

// Response wraps the streaming body and metadata needed by the HTTP layer.
type Response struct {
	Body          io.ReadCloser
	Headers       http.Header
	Status        int
	ContentLength int64
}

// Close closes the underlying response body if present.
func (r *Response) Close() error {
	if r == nil || r.Body == nil {
		return nil
	}
	return r.Body.Close()
}

// Provider supplies streaming data for a given request.
type Provider interface {
	Stream(ctx context.Context, req Request) (*Response, error)
}
