package tests

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/ThomasFerro/gogetter/app"
)

type SubstitutedRequest struct {
	app.Request
	Response     string
	ResponseCode int
}

func (s SubstitutedRequest) Apply(t TestHttpClient) TestHttpClient {
	t.SubstitutedRequests = append(t.SubstitutedRequests, s)
	return t
}

func (s SubstitutedRequest) ToHttpResponse() *http.Response {
	return &http.Response{
		Status:     "200 OK",
		StatusCode: s.ResponseCode,
		Body:       io.NopCloser(strings.NewReader(s.Response)),
	}
}

type TestHttpClient struct {
	SubstitutedRequests []SubstitutedRequest
}

func (t TestHttpClient) Do(req *http.Request) (*http.Response, error) {
	substitutedRequests, err := t.foundSubstitutedRequest(req)
	if err != nil {
		return nil, err
	}
	return substitutedRequests.ToHttpResponse(), nil
}

func (t TestHttpClient) foundSubstitutedRequest(req *http.Request) (SubstitutedRequest, error) {
	for _, substitutedRequest := range t.SubstitutedRequests {
		if substitutedRequest.Method != req.Method || substitutedRequest.Url != req.URL.String() {
			continue
		}

		allHeadersAreMatching := true
		for header, value := range substitutedRequest.Headers {
			requestHeader := req.Header[header]

			if len(requestHeader) != 1 || requestHeader[0] != value {
				allHeadersAreMatching = false
			}
		}
		if !allHeadersAreMatching {
			continue
		}

		return substitutedRequest, nil
	}
	return SubstitutedRequest{}, errors.New("request was not substituted")
}

type TestClientOption interface {
	Apply(t TestHttpClient) TestHttpClient
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
