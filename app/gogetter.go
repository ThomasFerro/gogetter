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

type History []Request

type Gogetter struct {
	client        HttpClient
	history       History
	historyWriter func([]byte) error
}

func (g Gogetter) History() History { return g.history }

func (g Gogetter) Execute(method, url string) (Gogetter, *http.Response, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return g, nil, fmt.Errorf("new request error: %w", err)
	}
	response, err := g.client.Do(req)
	if err != nil {
		return g, nil, fmt.Errorf("request execution error: %w", err)
	}

	responseCode := 0
	if response != nil {
		responseCode = response.StatusCode
	}
	g, err = g.AppendToHistory(method, url, responseCode)
	if err != nil {
		return g, nil, fmt.Errorf("unable to append to history: %w", err)
	}
	return g, response, nil
}

func (g Gogetter) AppendToHistory(method, url string, responseCode int) (Gogetter, error) {
	g.history = append(g.history, Request{
		Url:          url,
		Method:       method,
		ResponseCode: responseCode,
	})
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

// TODO: Stream the history ?
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
