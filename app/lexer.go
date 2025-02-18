package app

import (
	"strings"
)

var ignoredChars = []string{" ", "\t", "\r", "\n"}

const (
	LITERAL      tokenType = "LITERAL"
	HEADER       tokenType = "=:"
	SEARCH_PARAM tokenType = "=?"
	FORM_DATA    tokenType = "="
)

type tokenType string

type token struct {
	Type    tokenType
	Literal string
}

type lexer struct {
	input        string
	tokens       []token
	readPosition int
}

func (l lexer) peak(startingPosition, numberOfCharToEat int) string {
	endPosition := startingPosition + numberOfCharToEat
	if endPosition >= len(l.input) {
		endPosition = len(l.input) - 1
	}

	return l.input[startingPosition:endPosition]
}

func (l lexer) readUntilNextSeparator() string {
	subInput := l.input[l.readPosition:]
	closestSeparatorIndex := len(subInput)
	separators := append(ignoredChars, string(HEADER), string(SEARCH_PARAM))
	for _, separator := range separators {
		index := strings.Index(subInput, separator)
		if index != -1 && index < closestSeparatorIndex {
			closestSeparatorIndex = index
		}
	}
	if closestSeparatorIndex != -1 {
		return subInput[:closestSeparatorIndex]
	}
	return subInput

}

func (l lexer) nextCharIsToIgnore() bool {
	nextChar := string(l.input[l.readPosition])
	for _, separator := range ignoredChars {
		if separator == nextChar {
			return true
		}
	}
	return false
}

func (l lexer) skipIgnoredChars() lexer {
	for l.nextCharIsToIgnore() {
		l.readPosition += 1
	}
	return l
}

func (l lexer) next() (lexer, bool) {
	if l.readPosition >= len(l.input)-1 {
		return l, true
	}
	l = l.skipIgnoredChars()

	nextChar := string(l.input[l.readPosition])
	readPositionIncrement := 1

	switch nextChar {
	case "=":
		potentialParameter := l.peak(l.readPosition, 2)
		if potentialParameter == string(HEADER) {
			l.tokens = append(l.tokens, token{
				Type:    HEADER,
				Literal: potentialParameter,
			})
			readPositionIncrement = 2
			break
		}
		if potentialParameter == string(SEARCH_PARAM) {
			l.tokens = append(l.tokens, token{
				Type:    SEARCH_PARAM,
				Literal: potentialParameter,
			})
			readPositionIncrement = 2
			break
		}
		literal := l.readUntilNextSeparator()
		readPositionIncrement = len(literal)
		l.tokens = append(l.tokens, token{
			Type:    LITERAL,
			Literal: literal,
		})
	default:
		literal := l.readUntilNextSeparator()
		readPositionIncrement = len(literal)
		l.tokens = append(l.tokens, token{
			Type:    LITERAL,
			Literal: literal,
		})
	}

	l.readPosition += readPositionIncrement
	return l, false
}

func lexicalAnalysis(input string) lexer {
	lexer := lexer{
		input:        input,
		readPosition: 0,
		tokens:       []token{},
	}

	eof := false
	for !eof {
		lexer, eof = lexer.next()
	}

	return lexer
}
