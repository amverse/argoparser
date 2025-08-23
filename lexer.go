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
		NewTokenType tokenType
		AppendWith   rune
		ShouldFlush  bool
	}

	moveTo := func(params moveToParams) {
		state = params.NewState
		if params.AppendWith != 0 {
			currentToken.Value.WriteRune(params.AppendWith)
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
					NewState:   stateMetHyphen,
					AppendWith: runeSlice[pos],
				})
			} else if unicode.IsSpace(runeSlice[pos]) {
				continue
			} else if isQuote(runeSlice[pos]) {
				moveTo(moveToParams{
					NewState:     stateReadingQuotedString,
					NewTokenType: typeStringValue,
				})
				openedQuote = runeSlice[pos]
			} else {
				moveTo(moveToParams{
					NewState:     stateReadingSimpleStrign,
					NewTokenType: typeStringValue,
					AppendWith:   runeSlice[pos],
				})
			}
		case stateMetHyphen:
			if runeSlice[pos] == '-' {
				moveTo(moveToParams{
					NewState:     stateReadingLongKey,
					NewTokenType: typeLongKey,
					AppendWith:   runeSlice[pos],
				})
			} else if unicode.IsSpace(runeSlice[pos]) {
				moveTo(moveToParams{
					NewState:    stateInitial,
					ShouldFlush: true,
				})
			} else {
				moveTo(moveToParams{
					NewState:     stateReadingShortGroup,
					NewTokenType: typeShortGroup,
					AppendWith:   runeSlice[pos],
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
					NewState:   stateReadingShortGroup,
					AppendWith: runeSlice[pos],
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
					NewState:   stateReadingLongKey,
					AppendWith: runeSlice[pos],
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
					NewState:   stateReadingSimpleStrign,
					AppendWith: runeSlice[pos],
				})
			}
		case stateReadingQuotedString:
			if runeSlice[pos] == openedQuote {
				moveTo(moveToParams{
					NewState:    stateInitial,
					ShouldFlush: true,
				})
			} else if runeSlice[pos] == '\\' {
				moveTo(moveToParams{
					NewState: stateEscaped,
				})
			} else {
				moveTo(moveToParams{
					NewState:   stateReadingQuotedString,
					AppendWith: runeSlice[pos],
				})
			}
		case stateEscaped:
			if runeSlice[pos] == 'r' {
				moveTo(moveToParams{
					NewState:   stateReadingQuotedString,
					AppendWith: rune('\r'),
				})
			} else if runeSlice[pos] == 'n' {
				moveTo(moveToParams{
					NewState:   stateReadingQuotedString,
					AppendWith: rune('\n'),
				})
			} else if runeSlice[pos] == 't' {
				moveTo(moveToParams{
					NewState:   stateReadingQuotedString,
					AppendWith: rune('\t'),
				})
			} else {
				moveTo(moveToParams{
					NewState:   stateReadingQuotedString,
					AppendWith: runeSlice[pos],
				})
			}
		}
	}

	if currentToken.TokenType != 0 && currentToken.Value.Len() > 0 {
		flushToken()
	}

	return result
}
