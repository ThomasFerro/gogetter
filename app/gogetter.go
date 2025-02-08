package app

import (
	"fmt"
	"net/http"
)

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Request struct {
	Method       string
	Url          string
	ResponseCode int
}

func (r Request) FilterValue() string { return fmt.Sprintf("%v %v", r.Method, r.Url) }
func (r Request) String() string      { return fmt.Sprintf("[%v]%v (%v)", r.Method, r.Url, r.ResponseCode) }

type Gogetter struct {
	client        HttpClient
	history       History
	historyWriter func([]byte) error
}

func (g Gogetter) History() History { return g.history }

func (g Gogetter) Execute(method, url string) (Gogetter, Request, *http.Response, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return g, Request{}, nil, fmt.Errorf("new request error: %w", err)
	}
	response, err := g.client.Do(req)
	if err != nil {
		return g, Request{}, nil, fmt.Errorf("request execution error: %w", err)
	}

	responseCode := 0
	if response != nil {
		responseCode = response.StatusCode
	}
	request := Request{
		Url:          url,
		Method:       method,
		ResponseCode: responseCode,
	}
	g, err = g.AppendToHistory(request)
	if err != nil {
		return g, request, nil, fmt.Errorf("unable to append to history: %w", err)
	}
	return g, request, response, nil
}

type GogetterOption interface {
	Apply(Gogetter) (Gogetter, error)
}

func NewGogetter(client HttpClient, options ...GogetterOption) (Gogetter, error) {
	gogetter := Gogetter{
		client: client,
	}
	for _, option := range options {
		var err error
		gogetter, err = option.Apply(gogetter)
		if err != nil {
			return Gogetter{}, err
		}
	}

	return gogetter, nil
}
