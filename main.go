package main

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/ThomasFerro/gogetter/app"
	"github.com/ThomasFerro/gogetter/helpers"
	"github.com/ThomasFerro/gogetter/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func optionFileReader(filename string) (io.ReadCloser, error) {
	historyFileReader, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return helpers.EmptyReadCloser{}, nil
		}
		return nil, err
	}
	return historyFileReader, nil
}

const historyFilename = "gogetter_history.json"

func historyOption() (app.WithHistory, io.ReadCloser, error) {
	historyFileReader, err := optionFileReader(historyFilename)
	if err != nil {
		return app.WithHistory{}, nil, errors.New("history file reader error")
	}

	historyWritingFunc := func(toWrite []byte) error {
		historyFileWriter, err := os.Create(historyFilename)
		if err != nil {
			return err
		}
		defer historyFileWriter.Close()
		_, err = historyFileWriter.Write(toWrite)
		return err
	}
	return app.WithHistory{
		PreviousHistory: historyFileReader, HistoryWriter: historyWritingFunc,
	}, historyFileReader, nil
}

const savedRequestsFilename = "gogetter_requests.json"

func savedRequestsOption() (app.WithSavedRequests, io.ReadCloser, error) {
	savedRequestsFileReader, err := optionFileReader(savedRequestsFilename)
	if err != nil {
		return app.WithSavedRequests{}, nil, errors.New("saved requests file reader error")
	}

	requestsSavingFunc := func(toWrite []byte) error {
		savedRequestsFileWriter, err := os.Create(savedRequestsFilename)
		if err != nil {
			return err
		}
		defer savedRequestsFileWriter.Close()
		// FIXME: suffisant pour tout écrire ou il faut boucler ? + mutualiser ou utiliser un helper standard
		_, err = savedRequestsFileWriter.Write(toWrite)
		return err
	}
	return app.WithSavedRequests{
		InitialSavedRequests: savedRequestsFileReader, RequestsSavingFunc: requestsSavingFunc,
	}, savedRequestsFileReader, nil
}

func main() {
	withHistory, historyFileReader, err := historyOption()
	defer historyFileReader.Close()
	if err != nil {
		slog.Error("error while creating history option", slog.Any("error", err))
		os.Exit(1)
	}
	withSavedRequests, savedRequestsFileReader, err := savedRequestsOption()
	defer savedRequestsFileReader.Close()
	if err != nil {
		slog.Error("error while creating saved requests option", slog.Any("error", err))
		os.Exit(1)
	}
	gogetter, err := app.NewGogetter(http.DefaultClient, withHistory, withSavedRequests)
	if err != nil {
		slog.Error("error while creating new gogetter", slog.Any("error", err))
		os.Exit(1)
	}
	if _, err = tea.NewProgram(tui.NewModel(gogetter), tea.WithAltScreen()).Run(); err != nil {
		slog.Error("error while running program", slog.Any("error", err))
		os.Exit(1)
	}
}
