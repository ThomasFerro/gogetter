package main

import (
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

func main() {
	historyFileReader, err := historyFileReader()
	if err != nil {
		slog.Error("error while opening history file", slog.Any("error", err))
		os.Exit(1)
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
	gogetter, err := app.NewGogetter(http.DefaultClient, historyFileReader, historyWritingFunc)
	if err != nil {
		slog.Error("error while creating new gogetter", slog.Any("error", err))
		os.Exit(1)
	}
	if _, err = tea.NewProgram(tui.NewModel(gogetter), tea.WithAltScreen()).Run(); err != nil {
		slog.Error("error while running program", slog.Any("error", err))
		os.Exit(1)
	}
}
