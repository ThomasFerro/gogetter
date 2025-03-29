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

// TODO: Send request with a json body

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
Accept=:text/html x-api-key=:myApiKey`

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

func TestShouldSendARequestWithQueryParams(t *testing.T) {
	gogetter := tests.NewTestSetup(
		t,
		tests.SubstitutedRequest{
			Request: app.Request{
				Method: "GET", Url: "https://pkg.go.dev?a=b&orderBy=name&search=http+template&tag=standard+library"},
			Response: "ok"},
	)
	var result *http.Response
	var err error
	rawRequest := `GET https://pkg.go.dev?search=http%20template&orderBy=date
orderBy=?name tag=?"standard library" a=?b`

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

// TODO: Multiple values for same key
// TODO: Only one body element is sent ?
func TestShouldSendARequestWithMultipartBody(t *testing.T) {
	gogetter := tests.NewTestSetup(
		t,
		tests.SubstitutedRequest{
			Request: app.Request{
				Method: "GET",
				Url:    "https://pkg.go.dev",
				MultipartBody: app.MultipartBody{
					"key":       "value",
					"secondKey": "second value",
				},
			},
			Response:     "ok",
			ResponseCode: 200,
		},
	)
	var result *http.Response
	var err error
	rawRequest := `GET https://pkg.go.dev
key=value secondKey="second value"`

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
