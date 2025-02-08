package tests

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/ThomasFerro/gogetter/app"
)

type request struct {
	Method         string
	Url            string
	StringResponse string
	ResponseCode   int
}

func (r request) Response() *http.Response {
	return &http.Response{
		Status:     "200 OK",
		StatusCode: r.ResponseCode,
		Body:       io.NopCloser(strings.NewReader(r.StringResponse)),
	}
}

type TestHttpClient struct {
	SubstitutedRequests []request
}

func (t TestHttpClient) Do(req *http.Request) (*http.Response, error) {
	substitutedRequests, err := t.foundSubstitutedRequest(req)
	if err != nil {
		return nil, err
	}
	return substitutedRequests.Response(), nil
}

func (t TestHttpClient) foundSubstitutedRequest(req *http.Request) (request, error) {
	for _, substitutedRequest := range t.SubstitutedRequests {
		if substitutedRequest.Method == req.Method && substitutedRequest.Url == req.URL.String() {
			return substitutedRequest, nil
		}
	}
	return request{}, errors.New("request was not substituted")
}

type TestClientOption interface {
	Apply(t TestHttpClient) TestHttpClient
}

type SubstitutedRequestOption struct {
	Method       string
	Url          string
	Response     string
	ResponseCode int
}

func (s SubstitutedRequestOption) Apply(t TestHttpClient) TestHttpClient {
	t.SubstitutedRequests = append(t.SubstitutedRequests, request{
		Method:         s.Method,
		Url:            s.Url,
		StringResponse: s.Response,
		ResponseCode:   s.ResponseCode,
	})
	return t
}

func NewTestClient(options ...TestClientOption) TestHttpClient {
	testClient := TestHttpClient{}

	for _, option := range options {
		testClient = option.Apply(testClient)
	}

	return testClient
}

func NewTestSetup(t *testing.T, options ...TestClientOption) app.Gogetter {
	testClient := NewTestClient(options...)
	gogetter, err := app.NewGogetter(
		testClient,
		app.WithHistory{
			PreviousHistory: strings.NewReader("[]"),
			HistoryWriter:   func(toWrite []byte) error { return nil }})
	if err != nil {
		t.Fatalf("new gogetter failed: %v", err)
	}
	return gogetter
}
