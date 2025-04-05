package tests_test

import (
	"os"
	"strings"
	"testing"

	"github.com/ThomasFerro/gogetter/app"
	"github.com/ThomasFerro/gogetter/helpers"
	"github.com/ThomasFerro/gogetter/tests"
)

func TestShouldPutRequestsInLocalHistory(t *testing.T) {
	gogetter := tests.NewTestSetup(
		t,
		tests.SubstitutedRequest{Request: app.Request{Method: "GET", Url: "https://pkg.go.dev"}, Response: "ok"},
	)
	var err error
	gogetter, _, _, err = gogetter.Execute(app.Request{Method: "GET", Url: "https://pkg.go.dev"})
	if err != nil {
		t.Fatalf("request execution failed: %v", err)
	}

	history := gogetter.History()
	if len(history) != 1 {
		t.Fatalf("history not filled correctly: %v", history)
	}
}

func TestShouldLoadAndPersistHistory(t *testing.T) {
	firstRequest, err := app.ParseRequest("POST https://pkg.go.dev key=value X-Api-Key=:api-key")
	if err != nil {
		t.Fatalf("request parsing failed: %v", err)
	}
	secondRequest, err := app.ParseRequest("DELETE https://pkg.go.dev/1")
	if err != nil {
		t.Fatalf("request parsing failed: %v", err)
	}
	testClient := tests.NewTestClient(
		tests.SubstitutedRequest{
			Request: firstRequest, Response: "created", ResponseCode: 201},
		tests.SubstitutedRequest{Request: secondRequest, Response: "not authorized", ResponseCode: 401},
	)
	previousHistory := strings.NewReader(`[{ "request": "GET https:/google.fr", "responseCode": 200 }]`)
	historyFile, err := os.CreateTemp("", "test_*")
	if err != nil {
		t.Fatalf("history file creation failed: %v", err)
	}
	historyFileName := historyFile.Name()
	historyFile.Close()

	historyWritingFunc := func(toWrite []byte) error {
		historyFileWriter, err := os.Create(historyFileName)
		if err != nil {
			return err
		}
		defer historyFileWriter.Close()
		_, err = historyFileWriter.Write(toWrite)
		return err
	}

	gogetter, err := app.NewGogetter(testClient, app.WithHistory{PreviousHistory: previousHistory, HistoryWriter: historyWritingFunc})
	if err != nil {
		t.Fatalf("new gogetter failed: %v", err)
	}
	history := gogetter.History()
	if len(history) != 1 ||
		history[0].Method != "GET" ||
		history[0].Url != "https:/google.fr" ||
		history[0].ResponseCode != 200 {
		t.Fatalf("history not initially filled correctly: %v", history)
	}

	gogetter, _, _, err = gogetter.Execute(firstRequest)
	if err != nil {
		t.Fatalf("request execution failed: %v", err)
	}
	gogetter, _, _, err = gogetter.Execute(secondRequest)
	if err != nil {
		t.Fatalf("request execution failed: %v", err)
	}

	history = gogetter.History()
	if len(history) != 3 ||
		history[1].Method != "POST" ||
		history[1].Url != "https://pkg.go.dev" ||
		history[1].ResponseCode != 201 ||
		history[1].Headers["X-Api-Key"] != "api-key" ||
		history[1].MultipartBody["key"] != "value" ||
		history[2].Method != "DELETE" ||
		history[2].Url != "https://pkg.go.dev/1" ||
		history[2].ResponseCode != 401 {
		t.Fatalf("history not filled correctly: %v", history)
	}

	expectedWrite := `[{"Request":"GET https:/google.fr","ResponseCode":200},{"Request":"POST https://pkg.go.dev key=value X-Api-Key=:api-key","ResponseCode":201},{"Request":"DELETE https://pkg.go.dev/1","ResponseCode":401}]`
	actualHistory, err := os.ReadFile(historyFileName)
	if err != nil {
		t.Fatalf("history file reading failed: %v", err)
	}
	if string(actualHistory) != expectedWrite {
		t.Fatalf("history not wrote correctly: %v", string(actualHistory))
	}
}

func TestEmptyHistory(t *testing.T) {
	testClient := tests.NewTestClient()
	previousHistory := helpers.EmptyReadCloser{}
	historyWritingFunc := func(toWrite []byte) error { return nil }

	gogetter, err := app.NewGogetter(testClient, app.WithHistory{PreviousHistory: previousHistory, HistoryWriter: historyWritingFunc})
	if err != nil {
		t.Fatalf("new gogetter failed: %v", err)
	}
	history := gogetter.History()
	if len(history) != 0 {

		t.Fatalf("expected an empty history: %s", history)
	}
}

func TestNoHistoryOption(t *testing.T) {
	testClient := tests.NewTestClient(
		tests.SubstitutedRequest{Request: app.Request{Method: "GET", Url: "https://pkg.go.dev"}, Response: "ok"},
	)
	gogetter, err := app.NewGogetter(testClient)
	if err != nil {
		t.Fatalf("new gogetter failed: %v", err)
	}

	gogetter, _, _, err = gogetter.Execute(app.Request{Method: "GET", Url: "https://pkg.go.dev"})
	if err != nil {
		t.Fatalf("request execution failed: %v", err)
	}

	history := gogetter.History()
	if len(history) != 1 {
		t.Fatalf("in memory history not filled correctly: %v", history)
	}
}
