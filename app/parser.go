package app

import (
	"errors"
	"fmt"
	"slices"
)

var availableMethods = []string{"GET", "POST", "PUT", "DELETE"}

func extractMethod(lexer lexer) (string, error) {
	if slices.Index(availableMethods, lexer.tokens[0].Literal) == -1 {
		return "", errors.New("invalid request, provide a valid method")
	}
	return lexer.tokens[0].Literal, nil
}

func extractNextValue(requestAdditionalParameters []token, index int) (string, int) {
	if index >= len(requestAdditionalParameters)-1 {
		return "", index + 1
	}
	valueIndex := index + 1
	literal := requestAdditionalParameters[valueIndex].Literal
	if literal != string(QUOTES) {
		return literal, valueIndex
	}

	valueInsideQuotes := ""
	for {
		valueIndex++

		literal = requestAdditionalParameters[valueIndex].Literal
		if literal == string(QUOTES) {
			break
		}

		valueInsideQuotes = fmt.Sprintf("%v%v ", valueInsideQuotes, literal)

		if valueIndex >= len(requestAdditionalParameters)-1 {
			break
		}
	}

	return valueInsideQuotes[:len(valueInsideQuotes)-1], valueIndex
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
	index := 0
	for index < len(requestAdditionalParameters) {
		token := requestAdditionalParameters[index]
		if index == 0 || index >= len(requestAdditionalParameters) {
			index++
			continue
		}
		if token.Type == HEADER {
			header := requestAdditionalParameters[index-1].Literal
			value := ""
			value, index = extractNextValue(requestAdditionalParameters, index)
			request.Headers[header] = value
			continue
		}
		if token.Type == SEARCH_PARAM {
			key := requestAdditionalParameters[index-1].Literal
			value := ""
			value, index = extractNextValue(requestAdditionalParameters, index)
			request.SearchParams[key] = value
			continue
		}
		index++
	}

	return request, nil
}
