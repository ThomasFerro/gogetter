package app

import (
	"strings"
)

var ignoredChars = []string{" ", "\t", "\r", "\n"}

type token struct {
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
	input := l.input[l.readPosition:]
	nextCharIndex := 0
	for nextCharIndex < len(input) {
		nextChar := input[nextCharIndex]
		nextCharIndex = nextCharIndex + eatEverythingInsideQuotes(input[nextCharIndex:])
		for _, ignoredChar := range ignoredChars {
			if string(nextChar) == ignoredChar {
				return l.input[l.readPosition : l.readPosition+nextCharIndex]
			}
		}
		nextCharIndex++
	}

	return l.input[l.readPosition:]
}

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

func (l lexer) nextCharIsToIgnore() bool {
	if l.readPosition >= len(l.input) {
		return false
	}
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
	l = l.skipIgnoredChars()
	if l.readPosition >= len(l.input) {
		return l, true
	}

	literal := l.readUntilNextSeparator()
	readPositionIncrement := len(literal)
	l.tokens = append(l.tokens, token{
		Literal: literal,
	})

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
