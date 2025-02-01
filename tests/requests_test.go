package tests_test

import (
	"io"
	"testing"

	"github.com/ThomasFerro/gogetter/app"
	"github.com/ThomasFerro/gogetter/tests"
)

func TestShouldSendSimpleRequest(t *testing.T) {
	testClient := tests.NewTestClient(
		tests.SubstitutedRequestOption{Method: "GET", Url: "https://pkg.go.dev", Response: "ok"},
	)
	gogetter := app.NewGogetter(testClient)
	result, err := gogetter.Execute("GET", "https://pkg.go.dev")
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

