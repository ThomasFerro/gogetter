package tests_test

import (
	"os"
	"strings"
	"testing"

	"github.com/ThomasFerro/gogetter/app"
	"github.com/ThomasFerro/gogetter/helpers"
	"github.com/ThomasFerro/gogetter/tests"
)

func TestShouldProvideSavedRequests(t *testing.T) {
	testClient := tests.NewTestClient()
	initialSavedRequests := strings.NewReader(`[{ "method": "GET", "url": "https://pkg.go.dev" },{ "method": "POST", "url": "https:/my-api.com/posts" }]`)
	gogetter, err := app.NewGogetter(testClient, app.WithSavedRequests{InitialSavedRequests: initialSavedRequests, RequestsSavingFunc: nil})
	if err != nil {
		t.Fatalf("new gogetter failed: %v", err)
	}

	savedRequests := gogetter.SavedRequests()
	if len(savedRequests) != 2 ||
		savedRequests[0].Method != "GET" ||
		savedRequests[0].Url != "https://pkg.go.dev" ||
		savedRequests[1].Method != "POST" ||
		savedRequests[1].Url != "https:/my-api.com/posts" {
		t.Fatalf("saved requests not filled correctly: %v", savedRequests)
	}
}

func TestShouldSaveRequest(t *testing.T) {
	testClient := tests.NewTestClient()
	initialSavedRequests := strings.NewReader(`[{ "method": "GET", "url": "https://pkg.go.dev" }]`)
	savedRequestsFile, err := os.CreateTemp("", "test_*")

	if err != nil {
		t.Fatalf("saved requests file creation failed: %v", err)
	}
	savedRequestsFileName := savedRequestsFile.Name()
	savedRequestsFile.Close()

	requestsSavingFunc := func(toWrite []byte) error {
		savedRequestsFileWriter, err := os.Create(savedRequestsFileName)
		if err != nil {
			return err
		}
		defer savedRequestsFileWriter.Close()
		_, err = savedRequestsFileWriter.Write(toWrite)
		return err
	}

	gogetter, err := app.NewGogetter(testClient, app.WithSavedRequests{InitialSavedRequests: initialSavedRequests, RequestsSavingFunc: requestsSavingFunc})
	if err != nil {
		t.Fatalf("new gogetter failed: %v", err)
	}

	gogetter, err = gogetter.SaveRequest(app.Request{
		Method: "POST",
		Url:    "https:/my-api.com/posts",
	})
	if err != nil {
		t.Fatalf("request saving failed: %v", err)
	}

	savedRequests := gogetter.SavedRequests()
	if len(savedRequests) != 2 ||
		savedRequests[0].Method != "GET" ||
		savedRequests[0].Url != "https://pkg.go.dev" ||
		savedRequests[1].Method != "POST" ||
		savedRequests[1].Url != "https:/my-api.com/posts" {
		t.Fatalf("in memory saved requests not filled correctly: %v", savedRequests)
	}
	expectedWrite := `[{"Method":"GET","Url":"https://pkg.go.dev"},{"Method":"POST","Url":"https:/my-api.com/posts"}]`
	actualSavedRequests, err := os.ReadFile(savedRequestsFileName)
	if err != nil {
		t.Fatalf("saved requests file reading failed: %v", err)
	}
	if string(actualSavedRequests) != expectedWrite {
		t.Fatalf("saved requests not wrote correctly: %v", string(actualSavedRequests))
	}
}

func TestShouldWorkWithoutSavedRequestOption(t *testing.T) {
	testClient := tests.NewTestClient(
		tests.SubstitutedRequestOption{Method: "GET", Url: "https://pkg.go.dev", Response: "ok"},
	)
	gogetter, err := app.NewGogetter(testClient)
	if err != nil {
		t.Fatalf("new gogetter failed: %v", err)
	}

	_, _, _, err = gogetter.Execute("GET", "https://pkg.go.dev")
	if err != nil {
		t.Fatalf("request execution failed: %v", err)
	}
}

func TestShouldWorkWithoutSavedRequestYet(t *testing.T) {
	testClient := tests.NewTestClient()

	initialSavedRequests := helpers.EmptyReadCloser{}
	requestsSavingFunc := func(toWrite []byte) error { return nil }

	_, err := app.NewGogetter(testClient, app.WithSavedRequests{InitialSavedRequests: initialSavedRequests, RequestsSavingFunc: requestsSavingFunc})
	if err != nil {
		t.Fatalf("new gogetter failed: %v", err)
	}
}
