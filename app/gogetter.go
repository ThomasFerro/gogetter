package app

import (
	"fmt"
	"net/http"
)

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Headers map[string]string
type SearchParams map[string]string

type Request struct {
	Raw          string
	Method       string
	Url          string
	Headers      Headers
	SearchParams SearchParams
}

func (r Request) FilterValue() string { return fmt.Sprintf("%v %v", r.Method, r.Url) }
func (r Request) String() string      { return fmt.Sprintf("[%v]%v", r.Method, r.Url) }

type RequestAndResponse struct {
	Request
	ResponseCode int
}

func (r RequestAndResponse) FilterValue() string { return fmt.Sprintf("%v %v", r.Method, r.Url) }
func (r RequestAndResponse) String() string {
	return fmt.Sprintf("[%v]%v (%v)", r.Method, r.Url, r.ResponseCode)
}

type Gogetter struct {
	client             HttpClient
	history            History
	historyWriter      func([]byte) error
	savedRequests      SavedRequests
	requestsSavingFunc func([]byte) error
}

func (g Gogetter) History() History             { return g.history }
func (g Gogetter) SavedRequests() SavedRequests { return g.savedRequests }

func (g Gogetter) Execute(request Request) (Gogetter, RequestAndResponse, *http.Response, error) {
	req, err := http.NewRequest(request.Method, request.Url, nil)
	for header, value := range request.Headers {
		req.Header.Add(header, value)
	}
	q := req.URL.Query()
	for key, value := range request.SearchParams {
		q.Set(key, value)
	}
	if len(request.SearchParams) > 0 {
		req.URL.RawQuery = q.Encode()
	}
	if err != nil {
		return g, RequestAndResponse{}, nil, fmt.Errorf("new request error: %w", err)
	}
	response, err := g.client.Do(req)
	if err != nil {
		return g, RequestAndResponse{}, nil, fmt.Errorf("request execution error: %w", err)
	}

	responseCode := 0
	if response != nil {
		responseCode = response.StatusCode
	}
	requestAndResponse := RequestAndResponse{
		Request:      request,
		ResponseCode: responseCode,
	}
	g, err = g.AppendToHistory(requestAndResponse)
	if err != nil {
		return g, requestAndResponse, nil, fmt.Errorf("unable to append to history: %w", err)
	}
	return g, requestAndResponse, response, nil
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
