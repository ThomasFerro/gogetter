package tests_test

import (
	"io"
	"net/http"
	"testing"

	"github.com/ThomasFerro/gogetter/app"
	"github.com/ThomasFerro/gogetter/tests"
)

func TestShouldSendSimpleRequest(t *testing.T) {
	gogetter := tests.NewTestSetup(
		t,
		tests.SubstitutedRequest{Request: app.Request{Method: "GET", Url: "https://pkg.go.dev"}, Response: "ok"},
	)
	var result *http.Response
	var err error
	gogetter, _, result, err = gogetter.Execute(app.Request{Method: "GET", Url: "https://pkg.go.dev"})
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

// TODO: Save request with all parameters
// TODO: History request with all parameters

// TODO: Send request with search params
// TODO: Send request with a json body
// TODO: Send request with a multipart body

func TestShouldSendARequestWithHeaders(t *testing.T) {
	gogetter := tests.NewTestSetup(
		t,
		tests.SubstitutedRequest{
			Request: app.Request{
				Method: "GET", Url: "https://pkg.go.dev", Headers: app.Headers{
					"X-Api-Key": "myApiKey",
					"Accept":    "text/html",
				}},
			Response: "ok"},
	)
	var result *http.Response
	var err error
	rawRequest := `GET https://pkg.go.dev
Accept=:text/html x-api-key=:myApiKey
`

	request, err := app.ParseRequest(rawRequest)
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
