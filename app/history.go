package app

import (
	"encoding/json"
	"fmt"
	"io"
)

type History []RequestAndResponse

func (g Gogetter) AppendToHistory(requestAndResponse RequestAndResponse) (Gogetter, error) {
	g.history = append(g.history, requestAndResponse)
	if g.historyWriter == nil {
		return g, nil
	}
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
