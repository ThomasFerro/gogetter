package app

import (
	"fmt"
	"net/http"
)

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Gogetter struct {
	client HttpClient
}

func (g Gogetter) Execute(method, url string) (*http.Response, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("new request error: %w", err)
	}
	response, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request execution error: %w", err)
	}
	return response, nil
}

func NewGogetter(client HttpClient) Gogetter {
	return Gogetter{
		client,
	}
}
