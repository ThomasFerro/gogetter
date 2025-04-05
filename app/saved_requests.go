package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

type SavedRequests []Request

type SavedRequestsWritingDto []string

func (g Gogetter) writeSavedRequests() error {
	savedRequests := SavedRequestsWritingDto{}

	for _, savedRequest := range g.savedRequests {
		savedRequests = append(savedRequests, savedRequest.Raw)
	}
	toWrite, err := json.Marshal(savedRequests)
	if err != nil {
		return fmt.Errorf("saved requests marshal error: %w", err)
	}
	err = g.requestsSavingFunc(toWrite)
	if err != nil {
		return fmt.Errorf("saved requests writing error: %w", err)
	}
	return nil
}

func (g Gogetter) SaveRequest(request Request) (Gogetter, error) {
	g.savedRequests = append(g.savedRequests, request)
	if g.requestsSavingFunc == nil {
		return g, nil
	}
	return g, g.writeSavedRequests()
}

func (g Gogetter) RemoveSavedRequest(index int) (Gogetter, error) {
	if index >= len(g.savedRequests) {
		return g, errors.New("cannot remove saved request, invalid index")
	}
	g.savedRequests = append(g.savedRequests[:index], g.savedRequests[index+1:]...)
	if g.requestsSavingFunc == nil {
		return g, nil
	}
	return g, g.writeSavedRequests()
}

func extractSavedRequests(reader io.Reader) (SavedRequests, error) {
	readerContent, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("saved requests reading error: %w", err)
	}
	if len(readerContent) == 0 {
		return SavedRequests{}, nil
	}
	var rawSavedRequests SavedRequestsWritingDto
	err = json.Unmarshal(readerContent, &rawSavedRequests)
	savedRequests := SavedRequests{}
	for _, rawSavedRequest := range rawSavedRequests {
		request, err := ParseRequest(rawSavedRequest)
		if err != nil {
			return nil, fmt.Errorf("saved request parsing error: %w", err)
		}
		savedRequests = append(savedRequests, request)
	}

	return savedRequests, nil
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
