package tests

import (
	"fmt"
	"io"
	"log/slog"
	"mime"
	"mime/multipart"
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

func (s SubstitutedRequest) String() string {
	return fmt.Sprintf("[%v] %v (headers: %v, search: %v, form: %v) => %v", s.Method, s.Url, s.Headers, s.SearchParams, s.MultipartBody, s.ResponseCode)
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

func extractJsonBody(req *http.Request) (string, error) {
	header := req.Header.Get("Content-Type")
	if header == "" {
		return "", nil
	}
	mediaType, _, err := mime.ParseMediaType(header)

	if err != nil {
		return "", fmt.Errorf("request mime parsing error: %w", err)
	}
	if mediaType != "application/json" || req.Body == nil {
		return "", nil
	}
	defer req.Body.Close()
	content, err := io.ReadAll(req.Body)
	if err != nil {
		return "", fmt.Errorf("request body read error: %w", err)
	}
	return string(content), nil
}

func extractMultipartForm(req *http.Request) (*multipart.Form, error) {
	header := req.Header.Get("Content-Type")
	if header == "" {
		return nil, nil
	}
	mediaType, _, err := mime.ParseMediaType(header)
	if err != nil {
		return nil, fmt.Errorf("request media type parsing error: %w", err)
	}
	if !strings.HasPrefix(mediaType, "multipart/") {
		return nil, nil
	}
	err = req.ParseMultipartForm(100_000)
	if err != nil {
		return nil, fmt.Errorf("request multipart form parsing error: %w", err)
	}
	return req.MultipartForm, nil
}

func (t TestHttpClient) foundSubstitutedRequest(req *http.Request) (SubstitutedRequest, error) {
	jsonBody, err := extractJsonBody(req)
	if err != nil {
		return SubstitutedRequest{}, fmt.Errorf("json body extraction error: %w", err)
	}
	multipartForm, err := extractMultipartForm(req)
	if err != nil {
		return SubstitutedRequest{}, fmt.Errorf("multipart form extraction error: %w", err)
	}
	for index, substitutedRequest := range t.SubstitutedRequests {
		if substitutedRequest.Method != req.Method || substitutedRequest.Url != req.URL.String() {
			slog.Info("substituted request not matching", slog.Any("index", index), slog.Any("substitute method", substitutedRequest.Method), slog.Any("substitute url", substitutedRequest.Url), slog.Any("request method", req.Method), slog.Any("request url", req.URL.String()))
			continue
		}

		allHeadersAreMatching := true
		for header, value := range substitutedRequest.Headers {
			requestHeader := req.Header[header]
			if len(requestHeader) != 1 || requestHeader[0] != value {
				slog.Info("substituted request not matching", slog.Any("index", index), slog.Any("header key", header), slog.Any("substitute header value", value), slog.Any("request header value", requestHeader))
				allHeadersAreMatching = false
			}
		}
		if !allHeadersAreMatching {
			continue
		}

		if jsonBody != "" && jsonBody != string(substitutedRequest.JsonBody) {
			slog.Info("substituted request not matching", slog.Any("index", index), slog.Any("expected json body", substitutedRequest.JsonBody), slog.Any("actual json body", jsonBody))
			continue
		}

		allMultipartFormElementsAreMatching := true
		if len(substitutedRequest.MultipartBody) > 0 && (multipartForm == nil || len(multipartForm.Value) == 0) {
			slog.Info("substituted request not matching", slog.Any("index", index), slog.Any("number of expected elements in multipart body", len(substitutedRequest.MultipartBody)))
			continue
		}
		for formElement, value := range substitutedRequest.MultipartBody {
			requestMultipartFormElement := multipartForm.Value[formElement]

			if len(requestMultipartFormElement) != 1 || requestMultipartFormElement[0] != value {
				slog.Info("substituted request not matching", slog.Any("index", index), slog.Any("form key", formElement), slog.Any("substitute form value", value), slog.Any("request form value", requestMultipartFormElement))
				allMultipartFormElementsAreMatching = false
			}
		}
		if !allMultipartFormElementsAreMatching {
			continue
		}

		return substitutedRequest, nil
	}
	return SubstitutedRequest{}, fmt.Errorf("request was not substituted, expected %v to be in %v", req.URL, t.SubstitutedRequests)
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
