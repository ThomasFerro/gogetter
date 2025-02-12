package parser

import (
	"errors"
	"slices"

	"github.com/ThomasFerro/gogetter/app"
)

var availableMethods = []string{"GET", "POST", "PUT", "DELETE"}

func extractMethod(lexer lexer) (string, error) {
	if slices.Index(availableMethods, lexer.tokens[0].Literal) == -1 {
		return "", errors.New("invalid request, provide a valid method")
	}
	return lexer.tokens[0].Literal, nil
}

func ParseRequest(input string) (app.Request, error) {
	lexer := lexicalAnalysis(input)
	request := app.Request{}
	if len(lexer.tokens) < 2 {
		return request, errors.New("invalid request, provide at least a method and the url")
	}

	method, err := extractMethod(lexer)
	if err != nil {
		return request, err
	}
	request.Method = method
	request.Url = lexer.tokens[1].Literal

	return request, nil
}
