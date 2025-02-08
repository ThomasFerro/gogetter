package app

import (
	"encoding/json"
	"fmt"
	"io"
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

type History []Request

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
		return g, Request{},nil, fmt.Errorf("request execution error: %w", err)
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

func (g Gogetter) AppendToHistory(request Request) (Gogetter, error) {
	g.history = append(g.history, request)
	toWrite, err := json.Marshal(g.history)
	if err != nil {
		return g, fmt.Errorf("history marshal error: %w", err)
	}
	err = g.historyWriter(toWrite)
	if err != nil {
		return g, fmt.Errorf("history writing error: %w", err)
	}
	return g, nil
}

func extractHistory(reader io.Reader) (History, error) {
	readerContent, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	if len(readerContent) == 0 {
		return History{}, nil
	}
	var history History
	err = json.Unmarshal(readerContent, &history)
	return history, err
}

func NewGogetter(client HttpClient, previousHistory io.Reader, historyWriter func([]byte) error) (Gogetter, error) {
	history, err := extractHistory(previousHistory)
	if err != nil {
		return Gogetter{}, err
	}
	return Gogetter{
		client:        client,
		history:       history,
		historyWriter: historyWriter,
	}, nil
}
