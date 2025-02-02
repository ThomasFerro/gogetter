package tests_test

import (
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/ThomasFerro/gogetter/app"
	"github.com/ThomasFerro/gogetter/tests"
)

func newTestSetup(t *testing.T, options ...tests.TestClientOption) app.Gogetter {
	testClient := tests.NewTestClient(options...)
	gogetter, err := app.NewGogetter(testClient, strings.NewReader("[]"), func(toWrite []byte) error { return nil })
	if err != nil {
		t.Fatalf("new gogetter failed: %v", err)
	}
	return gogetter
}

func TestShouldSendSimpleRequest(t *testing.T) {
	gogetter := newTestSetup(
		t,
		tests.SubstitutedRequestOption{Method: "GET", Url: "https://pkg.go.dev", Response: "ok"},
	)
	var result *http.Response
	var err error
	gogetter, result, err = gogetter.Execute("GET", "https://pkg.go.dev")
	if err != nil {
		t.Fatalf("request execution failed: %v", err)
	}
	if result == nil {
		t.Fatalf("no result")
	}
	defer result.Body.Close()
	body, err := io.ReadAll(result.Body)
	if err != nil {
		t.Fatalf("result body read failed: %v", err)
	}
	if string(body) != "ok" {
		t.Fatalf(`expected body to be "ok" but got %v`, string(body))
	}
}

func TestShouldPutRequestsInLocalHistory(t *testing.T) {
	gogetter := newTestSetup(
		t,
		tests.SubstitutedRequestOption{Method: "GET", Url: "https://pkg.go.dev", Response: "ok"},
	)
	var err error
	gogetter, _, err = gogetter.Execute("GET", "https://pkg.go.dev")
	if err != nil {
		t.Fatalf("request execution failed: %v", err)
	}

	history := gogetter.History()
	if len(history) != 1 {
		t.Fatalf("history not filled correctly: %v", history)
	}
}

func TestShouldLoadAndPersistHistory(t *testing.T) {
	testClient := tests.NewTestClient(
		tests.SubstitutedRequestOption{Method: "POST", Url: "https://pkg.go.dev", Response: "created", ResponseCode: 201},
		tests.SubstitutedRequestOption{Method: "DELETE", Url: "https://pkg.go.dev/1", Response: "not authorized", ResponseCode: 401},
	)
	previousHistory := strings.NewReader(`[{ "method": "GET", "url": "https:/google.fr", "responseCode": 200 }]`)
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

	gogetter, err := app.NewGogetter(testClient, previousHistory, historyWritingFunc)
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

	gogetter, _, err = gogetter.Execute("POST", "https://pkg.go.dev")
	if err != nil {
		t.Fatalf("request execution failed: %v", err)
	}
	gogetter, _, err = gogetter.Execute("DELETE", "https://pkg.go.dev/1")
	if err != nil {
		t.Fatalf("request execution failed: %v", err)
	}

	history = gogetter.History()
	if len(history) != 3 ||
		history[1].Method != "POST" ||
		history[1].Url != "https://pkg.go.dev" ||
		history[1].ResponseCode != 201 ||
		history[2].Method != "DELETE" ||
		history[2].Url != "https://pkg.go.dev/1" ||
		history[2].ResponseCode != 401 {
		t.Fatalf("history not filled correctly: %v", history)
	}

	expectedWrite := `[{"Method":"GET","Url":"https:/google.fr","ResponseCode":200},{"Method":"POST","Url":"https://pkg.go.dev","ResponseCode":201},{"Method":"DELETE","Url":"https://pkg.go.dev/1","ResponseCode":401}]`
	actualHistory, err := os.ReadFile(historyFileName)
	if err != nil {
		t.Fatalf("history file reading failed: %v", err)
	}
	if string(actualHistory) != expectedWrite {
		t.Fatalf("history not wrote correctly: %v", string(actualHistory))
	}
}
