package app

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"text/template"
)

var availableMethods = []string{"GET", "POST", "PUT", "DELETE"}

func extractMethod(firstInputElement string) (string, error) {
	if slices.Index(availableMethods, firstInputElement) == -1 {
		return "", errors.New("invalid request, provide a valid method")
	}
	return firstInputElement, nil
}

type keyword string

const (
	HEADER            keyword = "=:"
	SEARCH_PARAM      keyword = "=?"
	FORM_DATA         keyword = "="
	JSON_OBJECT_START keyword = "{"
	JSON_ARRAY_START  keyword = "["
)

func extractKeyValuePair(literal string, separator string) (key string, value string) {
	split := strings.Split(literal, separator)
	if strings.HasPrefix(split[1], "\"") && strings.HasSuffix(split[1], "\"") {
		return split[0], split[1][1 : len(split[1])-1]
	}
	return split[0], split[1]
}

var separators = []string{" ", "\t", "\r", "\n"}

func eatEverythingInsideQuotes(input string) (index int) {
	if string(input[0]) != "\"" {
		return 0
	}
	indexOfClosingQuote := strings.Index(input[1:], "\"")
	if indexOfClosingQuote != -1 {
		return indexOfClosingQuote + 1
	}
	return 0
}

func readUntilNextSeparator(input string) string {
	nextCharIndex := 0
	for nextCharIndex < len(input) {
		nextChar := input[nextCharIndex]
		nextCharIndex = nextCharIndex + eatEverythingInsideQuotes(input[nextCharIndex:])
		if slices.Contains(separators, string(nextChar)) {
			return input[:nextCharIndex]
		}
		nextCharIndex++
	}

	return input
}

func splitRequestInput(input string) []string {
	splitInput := []string{}
	readPosition := 0
	for readPosition < len(input) {
		restOfInput := input[readPosition:]
		inputElement := readUntilNextSeparator(restOfInput)
		splitInput = append(splitInput, inputElement)
		readPosition += len(inputElement) + 1
	}
	return splitInput
}

type RequestParsingOption interface {
	Apply(input string) (string, error)
}

type TemplatedRequestOption struct {
	Data any
}

func (o TemplatedRequestOption) Apply(input string) (string, error) {
	tmpl, err := template.New("request").Parse(input)
	if err != nil {
		return input, fmt.Errorf("template parsing error: %w", err)
	}
	builder := strings.Builder{}
	err = tmpl.Execute(&builder, o.Data)
	if err != nil {
		return input, err
	}
	return builder.String(), nil

}

func ParseRequest(input string, options ...RequestParsingOption) (Request, error) {
	var err error
	for _, option := range options {
		input, err = option.Apply(input)
		if err != nil {
			return Request{}, err
		}
	}

	request := Request{
		Raw: input,
	}
	inputElements := splitRequestInput(input)
	if len(inputElements) < 2 {
		return request, errors.New("invalid request, provide at least a method and the url")
	}

	method, err := extractMethod(inputElements[0])
	if err != nil {
		return request, err
	}
	request.Method = method
	request.Url = inputElements[1]
	request.Headers = Headers{}
	request.SearchParams = SearchParams{}
	request.MultipartBody = MultipartBody{}

	if len(inputElements) > 2 {
		requestAdditionalParameters := inputElements[2:]
		for _, additionalParameter := range requestAdditionalParameters {
			if strings.Contains(additionalParameter, string(HEADER)) {
				key, value := extractKeyValuePair(additionalParameter, string(HEADER))
				request.Headers[key] = value
				continue
			}
			if strings.Contains(additionalParameter, string(SEARCH_PARAM)) {
				key, value := extractKeyValuePair(additionalParameter, string(SEARCH_PARAM))
				request.SearchParams[key] = value
				continue
			}
			if strings.Contains(additionalParameter, string(FORM_DATA)) {
				key, value := extractKeyValuePair(additionalParameter, string(FORM_DATA))
				request.MultipartBody[key] = value
				continue
			}
			if strings.HasPrefix(additionalParameter, string(JSON_ARRAY_START)) {
				index := strings.Index(input, string(JSON_ARRAY_START))
				request.JsonBody = JsonBody(input[index:])
				break
			}
			if strings.HasPrefix(additionalParameter, string(JSON_OBJECT_START)) {
				index := strings.Index(input, string(JSON_OBJECT_START))
				request.JsonBody = JsonBody(input[index:])
				break
			}
		}
	}

	return request, nil
}
