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

	requestAdditionalParameters := lexer.tokens[2:]
	for index, token := range requestAdditionalParameters {
		if token.Type == HEADER && index > 0 && len(requestAdditionalParameters) > index+1 {
			header := requestAdditionalParameters[index-1].Literal
			value := requestAdditionalParameters[index+1].Literal
			request.Headers[header] = value
		}
	}

	return request, nil
}
