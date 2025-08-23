package argoparser

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
)

func isFlag(entry *indexEntry) bool {
	return entry.t.Kind() == reflect.Bool
}

func isMultiValue(entry *indexEntry) bool {
	return entry.t.Kind() == reflect.Slice
}

func consumeValue(entry *indexEntry, value string) error {
	castTo := entry.t
	if isMultiValue(entry) {
		castTo = entry.t.Elem()
	}

	var valueToAppend any
	var err error
	if castTo.Kind() == reflect.Int {
		valueToAppend, err = strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid value for int: %s", value)
		}
	} else if castTo.Kind() == reflect.String {
		valueToAppend = value
	} else {
		return fmt.Errorf("unsupported type: %s", castTo.Kind())
	}

	if isMultiValue(entry) {
		entry.v.Set(reflect.Append(entry.v, reflect.ValueOf(valueToAppend)))
	} else {
		entry.v.Set(reflect.ValueOf(valueToAppend))
	}

	entry.presented = true
	return nil
}

func checkRequiredFields(index fieldsIndex) error {
	for _, entry := range index.requiredFields {
		if !entry.presented {
			// TODO: change error text
			return fmt.Errorf("required field is not presented: %s", entry.m.longName)
		}
	}
	return nil
}

func parseImpl(tokens []token, result any) error {
	index, err := buildIndex(result)
	if err != nil {
		return err
	}

	tokenPos := 0

	positionalPos := 0

	for tokenPos < len(tokens) {
		token := tokens[tokenPos]

		switch token.TokenType {
		case typeLongKey:
			entry, ok := index.fieldsByLongName[token.Value]
			if !ok {
				return fmt.Errorf("unknown long key: %s", token.Value)
			}

			if isFlag(entry) {
				entry.v.SetBool(true)
				entry.presented = true
				tokenPos++
				continue
			}

			if tokenPos+1 >= len(tokens) {
				return fmt.Errorf("missing value for flag: %s", token.Value)
			}

			nextToken := tokens[tokenPos+1]

			if err := consumeValue(entry, nextToken.Value); err != nil {
				return err
			}

			tokenPos++
		case typeShortGroup:
			if len(token.Value) > 2 {
				flags := token.Value[1:]
				for _, flag := range flags {
					entry, ok := index.fieldsByShortName["-"+string(flag)]
					if !ok {
						return fmt.Errorf("unknown short key: %s", string(flag))
					}
					if !isFlag(entry) {
						return fmt.Errorf("value for field is flag, but field is not a flag: %s", "-"+string(flag))
					}
					entry.v.SetBool(true)
					entry.presented = true
				}
				tokenPos++
				continue
			}

			entry, ok := index.fieldsByShortName[token.Value]
			// the code below is copypasted from long-key parsing
			// TODO: move to common place
			if !ok {
				return fmt.Errorf("unknown short key: %s", token.Value)
			}

			if isFlag(entry) {
				entry.v.SetBool(true)
				entry.presented = true
				tokenPos++
				continue
			}

			if tokenPos+1 >= len(tokens) {
				return fmt.Errorf("missing value for flag: %s", token.Value)
			}

			nextToken := tokens[tokenPos+1]

			if err := consumeValue(entry, nextToken.Value); err != nil {
				return err
			}

			tokenPos++
		case typeStringValue:
			if positionalByIndex, ok := index.fieldsByIndex[positionalPos]; ok {
				if err := consumeValue(positionalByIndex, token.Value); err != nil {
					return err
				}
				positionalPos++
			} else {
				if index.positionalsDefault != nil {
					if err := consumeValue(index.positionalsDefault, token.Value); err != nil {
						return err
					}
				} else {
					return fmt.Errorf("unexpected positional parameter: %s", token.Value)
				}
			}
		}

		tokenPos++
	}

	if err := checkRequiredFields(index); err != nil {
		return err
	}

	return nil
}

func ParseString(input string, result any) error {
	tokens := lex(input)
	return parseImpl(tokens, result)
}

func ParseSlice(input []string, result any) error {
	return ParseString(strings.Join(input, " "), result)
}

func ParseAppArgs(result any) error {
	tokens := []token{}
	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "--") {
			tokens = append(tokens, token{TokenType: typeLongKey, Value: arg})
		} else if strings.HasPrefix(arg, "-") {
			tokens = append(tokens, token{TokenType: typeShortGroup, Value: arg})
		} else {
			tokens = append(tokens, token{TokenType: typeStringValue, Value: arg})
		}
	}

	return parseImpl(tokens, result)
}

func ParseReader(reader *io.Reader, result any) error {
	data, err := io.ReadAll(*reader)
	if err != nil {
		return err
	}

	return ParseString(string(data), result)
}
