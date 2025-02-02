package tests

import (
	"errors"
	"io"
	"net/http"
	"strings"
)

type request struct {
	Method         string
	Url            string
	StringResponse string
}

func (r request) Response() *http.Response {
	return &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
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
	Method   string
	Url      string
	Response string
}

func (s SubstitutedRequestOption) Apply(t TestHttpClient) TestHttpClient {
	t.SubstitutedRequests = append(t.SubstitutedRequests, request{
		Method:         s.Method,
		Url:            s.Url,
		StringResponse: s.Response,
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
