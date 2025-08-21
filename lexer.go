package argoparser

import (
	"strings"
	"unicode"
)

type tokenType int

const (
	_               tokenType = iota
	typeStringValue           // anything without hyphen: asdf, "asdf", 'asdf', `"asdf"` etc
	typeShortGroup            // -s, -abc
	typeLongKey               // --this-is-long-key
)

type token struct {
	TokenType tokenType
	Value     string
}

type tokenBuilder struct {
	TokenType tokenType
	Value     *strings.Builder
}

type lexerState int

const (
	stateInitial lexerState = iota
	stateMetHyphen
	stateReadingShortGroup
	stateReadingLongKey
	stateReadingSimpleStrign
	stateReadingQuotedString
	stateEscaped
)

func isQuote(r rune) bool {
	return r == '"' || r == '\'' || r == '`'
}

/**
* this is just a naive implementation of the automaton from lexerAutomaton.png
 */
func lex(input string) []token {
	result := make([]token, 0)

	runeSlice := []rune(input)

	state := stateInitial
	currentToken := tokenBuilder{
		Value: &strings.Builder{},
	}
	openedQuote := rune(0)

	appendCurrentValue := func(pos int) {
		currentToken.Value.WriteRune(runeSlice[pos])
	}

	flushToken := func() {
		result = append(result, token{
			TokenType: currentToken.TokenType,
			Value:     currentToken.Value.String(),
		})
		currentToken = tokenBuilder{
			Value: &strings.Builder{},
		}
	}

	type moveToParams struct {
		NewState     lexerState
		Pos          int
		NewTokenType tokenType
		ShouldAppend bool
		ShouldFlush  bool
	}

	moveTo := func(params moveToParams) {
		state = params.NewState
		if params.ShouldAppend {
			appendCurrentValue(params.Pos)
		}
		if params.NewTokenType != 0 {
			currentToken.TokenType = params.NewTokenType
		}
		if params.ShouldFlush {
			flushToken()
		}
	}

	for pos := 0; pos < len(runeSlice); pos++ {
		switch state {
		case stateInitial:
			if runeSlice[pos] == '-' {
				moveTo(moveToParams{
					NewState:     stateMetHyphen,
					Pos:          pos,
					ShouldAppend: true,
				})
			} else if unicode.IsSpace(runeSlice[pos]) {
				continue
			} else if isQuote(runeSlice[pos]) {
				moveTo(moveToParams{
					NewState:     stateReadingQuotedString,
					Pos:          pos,
					NewTokenType: typeStringValue,
				})
				openedQuote = runeSlice[pos]
			} else {
				moveTo(moveToParams{
					NewState:     stateReadingSimpleStrign,
					Pos:          pos,
					NewTokenType: typeStringValue,
					ShouldAppend: true,
				})
			}
		case stateMetHyphen:
			if runeSlice[pos] == '-' {
				moveTo(moveToParams{
					NewState:     stateReadingLongKey,
					Pos:          pos,
					NewTokenType: typeLongKey,
					ShouldAppend: true,
				})
			} else if unicode.IsSpace(runeSlice[pos]) {
				moveTo(moveToParams{
					NewState:    stateInitial,
					ShouldFlush: true,
				})
			} else {
				moveTo(moveToParams{
					NewState:     stateReadingShortGroup,
					Pos:          pos,
					NewTokenType: typeShortGroup,
					ShouldAppend: true,
				})
			}
		case stateReadingShortGroup:
			if unicode.IsSpace(runeSlice[pos]) {
				moveTo(moveToParams{
					NewState:    stateInitial,
					ShouldFlush: true,
				})
			} else {
				moveTo(moveToParams{
					NewState:     stateReadingShortGroup,
					Pos:          pos,
					ShouldAppend: true,
				})
			}
		case stateReadingLongKey:
			if unicode.IsSpace(runeSlice[pos]) {
				moveTo(moveToParams{
					NewState:    stateInitial,
					ShouldFlush: true,
				})
			} else {
				moveTo(moveToParams{
					NewState:     stateReadingLongKey,
					Pos:          pos,
					ShouldAppend: true,
				})
			}
		case stateReadingSimpleStrign:
			if unicode.IsSpace(runeSlice[pos]) {
				moveTo(moveToParams{
					NewState:    stateInitial,
					ShouldFlush: true,
				})
			} else {
				moveTo(moveToParams{
					NewState:     stateReadingSimpleStrign,
					Pos:          pos,
					ShouldAppend: true,
				})
			}
		case stateReadingQuotedString:
			if runeSlice[pos] == openedQuote {
				moveTo(moveToParams{
					NewState:    stateInitial,
					Pos:         pos,
					ShouldFlush: true,
				})
			} else if runeSlice[pos] == '\\' {
				moveTo(moveToParams{
					NewState: stateEscaped,
					Pos:      pos,
				})
			} else {
				moveTo(moveToParams{
					NewState:     stateReadingQuotedString,
					Pos:          pos,
					ShouldAppend: true,
				})
			}
		case stateEscaped:
			moveTo(moveToParams{
				NewState:     stateReadingQuotedString,
				Pos:          pos,
				ShouldAppend: true,
			})
		}
	}

	if currentToken.TokenType != 0 && currentToken.Value.Len() > 0 {
		flushToken()
	}

	return result
}
