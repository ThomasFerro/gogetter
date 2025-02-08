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

func historyFileReader() (io.ReadCloser, error) {
	historyFileReader, err := os.Open("gogetter_history.json")
	if err != nil {
		if os.IsNotExist(err) {
			return helpers.EmptyReadCloser{}, nil
		}
		return nil, err
	}
	return historyFileReader, nil
}

func historyOption() (app.WithHistory, error) {
	historyFileReader, err := historyFileReader()
	if err != nil {
		return app.WithHistory{}, errors.New("history file reader error")
	}
	defer historyFileReader.Close()

	historyWritingFunc := func(toWrite []byte) error {
		historyFileWriter, err := os.Create("gogetter_history.json")
		if err != nil {
			return err
		}
		defer historyFileWriter.Close()
		_, err = historyFileWriter.Write(toWrite)
		return err
	}
	return app.WithHistory{
		PreviousHistory: historyFileReader, HistoryWriter: historyWritingFunc,
	}, nil
}

func main() {
	withHistory, err := historyOption()
	if err != nil {
		slog.Error("error while creating history", slog.Any("error", err))
		os.Exit(1)
	}
	gogetter, err := app.NewGogetter(http.DefaultClient, withHistory)
	if err != nil {
		slog.Error("error while creating new gogetter", slog.Any("error", err))
		os.Exit(1)
	}
	if _, err = tea.NewProgram(tui.NewModel(gogetter), tea.WithAltScreen()).Run(); err != nil {
		slog.Error("error while running program", slog.Any("error", err))
		os.Exit(1)
	}
}
