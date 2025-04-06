package tests_test

import (
	"io"
	"net/http"
	"testing"

	"github.com/ThomasFerro/gogetter/app"
	"github.com/ThomasFerro/gogetter/tests"
)

func TestShouldSendTemplatedRequest(t *testing.T) {
	gogetter := tests.NewTestSetup(
		t,
		tests.SubstitutedRequest{Request: app.Request{Method: "GET", Url: "https://pkg.go.dev/123"}, Response: "ok"},
	)
	var result *http.Response
	var templateData = struct {
		Id string
	}{
		Id: "123",
	}
	templateOption := app.TemplatedRequestOption{
		Data: templateData,
	}
	request, err := app.ParseRequest("GET https://pkg.go.dev/{{.Id}}", templateOption)
	if err != nil {
		t.Fatalf("request parsing failed: %v", err)
	}
	gogetter, _, result, err = gogetter.Execute(request)
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
