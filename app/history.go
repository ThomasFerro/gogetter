package app

import (
	"encoding/json"
	"fmt"
	"io"
)

type HistoryEntryWritingDto struct {
	Request      string
	ResponseCode int
}

type History []RequestAndResponse

func (g Gogetter) AppendToHistory(requestAndResponse RequestAndResponse) (Gogetter, error) {
	g.history = append(g.history, requestAndResponse)
	if g.historyWriter == nil {
		return g, nil
	}

	history := []HistoryEntryWritingDto{}
	for _, request := range g.history {
		history = append(history, HistoryEntryWritingDto{
			Request:      request.Raw,
			ResponseCode: request.ResponseCode,
		})
	}
	toWrite, err := json.Marshal(history)
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
		return nil, fmt.Errorf("history reading error: %w", err)
	}
	if len(readerContent) == 0 {
		return History{}, nil
	}
	var rawHistory []HistoryEntryWritingDto
	err = json.Unmarshal(readerContent, &rawHistory)
	history := History{}
	for _, historyEntry := range rawHistory {
		request, err := ParseRequest(historyEntry.Request)
		if err != nil {
			return nil, fmt.Errorf("history entry parsing error: %w", err)
		}
		history = append(history, RequestAndResponse{
			Request:      request,
			ResponseCode: historyEntry.ResponseCode,
		})
	}

	return history, nil
}

type WithHistory struct {
	PreviousHistory io.Reader
	HistoryWriter   func([]byte) error
}

func (w WithHistory) Apply(g Gogetter) (Gogetter, error) {
	history, err := extractHistory(w.PreviousHistory)
	if err != nil {
		return Gogetter{}, err
	}
	g.history = history
	g.historyWriter = w.HistoryWriter
	return g, nil
}
