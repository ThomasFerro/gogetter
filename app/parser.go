package app

import (
	"errors"
	"slices"
	"strings"
)

var availableMethods = []string{"GET", "POST", "PUT", "DELETE"}

func extractMethod(lexer lexer) (string, error) {
	if slices.Index(availableMethods, lexer.tokens[0].Literal) == -1 {
		return "", errors.New("invalid request, provide a valid method")
	}
	return lexer.tokens[0].Literal, nil
}

type keyword string

const (
	HEADER       keyword = "=:"
	SEARCH_PARAM keyword = "=?"
	FORM_DATA    keyword = "="
)

func extractKeyValuePair(literal string, separator string) (key string, value string) {
	split := strings.Split(literal, separator)
	if strings.HasPrefix(split[1], "\"") && strings.HasSuffix(split[1], "\"") {
		return split[0], split[1][1 : len(split[1])-1]
	}
	return split[0], split[1]
}

func ParseRequest(input string) (Request, error) {
	lexer := lexicalAnalysis(input)
	request := Request{
		Raw: input,
	}
	if len(lexer.tokens) < 2 {
		return request, errors.New("invalid request, provide at least a method and the url")
	}

	method, err := extractMethod(lexer)
	if err != nil {
		return request, err
	}
	request.Method = method
	request.Url = lexer.tokens[1].Literal
	request.Headers = Headers{}
	request.SearchParams = SearchParams{}

	requestAdditionalParameters := lexer.tokens[2:]
	for _, token := range requestAdditionalParameters {
		if strings.Contains(token.Literal, string(HEADER)) {
			key, value := extractKeyValuePair(token.Literal, string(HEADER))
			request.Headers[key] = value
			continue
		}
		if strings.Contains(token.Literal, string(SEARCH_PARAM)) {
			key, value := extractKeyValuePair(token.Literal, string(SEARCH_PARAM))
			request.SearchParams[key] = value
			continue
		}
	}

	return request, nil
}
