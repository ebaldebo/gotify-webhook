package requester

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

type Requester struct {
	*http.Client
}

type HttpResponse struct {
	Status     string
	StatusCode int
	Header     http.Header
	Body       []byte
}

func NewRequester(client *http.Client) *Requester {
	return &Requester{
		Client: client,
	}
}

func (r *Requester) Post(ctx context.Context, url string, payload any, headers map[string]string) (*HttpResponse, error) {
	return r.SendRequest(ctx, http.MethodPost, url, payload, headers)
}

func (r *Requester) Get(ctx context.Context, url string, headers map[string]string) (*HttpResponse, error) {
	return r.SendRequest(ctx, http.MethodGet, url, nil, headers)
}

func (r *Requester) SendRequest(ctx context.Context, method, url string, payload any, headers map[string]string) (*HttpResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var body io.Reader
	if payload != nil {
		buffer, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}

		body = bytes.NewReader(buffer)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	if method == http.MethodPost {
		req.Header.Set("Content-Type", "application/json")
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	res, err := r.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	responseBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return &HttpResponse{
		Status:     res.Status,
		StatusCode: res.StatusCode,
		Header:     res.Header,
		Body:       responseBody,
	}, nil
}
