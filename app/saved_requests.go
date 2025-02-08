package app

import (
	"encoding/json"
	"fmt"
	"io"
)

type SavedRequests []Request

func (g Gogetter) SaveRequest(request Request) (Gogetter, error) {
	g.savedRequests = append(g.savedRequests, request)
	if g.requestsSavingFunc == nil {
		return g, nil
	}
	toWrite, err := json.Marshal(g.savedRequests)
	if err != nil {
		return g, fmt.Errorf("saved requests marshal error: %w", err)
	}
	err = g.requestsSavingFunc(toWrite)
	if err != nil {
		return g, fmt.Errorf("saved requests writing error: %w", err)
	}
	return g, nil
}

func extractSavedRequests(reader io.Reader) (SavedRequests, error) {
	readerContent, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	if len(readerContent) == 0 {
		return SavedRequests{}, nil
	}
	var savedRequests SavedRequests
	err = json.Unmarshal(readerContent, &savedRequests)
	return savedRequests, err
}

type WithSavedRequests struct {
	InitialSavedRequests io.Reader
	RequestsSavingFunc   func([]byte) error
}

func (w WithSavedRequests) Apply(g Gogetter) (Gogetter, error) {
	savedRequests, err := extractSavedRequests(w.InitialSavedRequests)
	if err != nil {
		return Gogetter{}, err
	}
	g.savedRequests = savedRequests
	g.requestsSavingFunc = w.RequestsSavingFunc
	return g, nil
}
