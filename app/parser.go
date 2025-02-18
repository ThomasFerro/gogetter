package app

import (
	"errors"
	"slices"
)

var availableMethods = []string{"GET", "POST", "PUT", "DELETE"}

func extractMethod(lexer lexer) (string, error) {
	if slices.Index(availableMethods, lexer.tokens[0].Literal) == -1 {
		return "", errors.New("invalid request, provide a valid method")
	}
	return lexer.tokens[0].Literal, nil
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
	for index, token := range requestAdditionalParameters {
		if index == 0 || index == len(requestAdditionalParameters)-1 {
			continue
		}
		if token.Type == HEADER {
			header := requestAdditionalParameters[index-1].Literal
			value := requestAdditionalParameters[index+1].Literal
			request.Headers[header] = value
			continue
		}
		if token.Type == SEARCH_PARAM {
			key := requestAdditionalParameters[index-1].Literal
			value := requestAdditionalParameters[index+1].Literal
			request.SearchParams[key] = value
			continue
		}
	}

	return request, nil
}
