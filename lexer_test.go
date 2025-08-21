package argoparser

import (
	"reflect"
	"testing"
)

func TestLexer(t *testing.T) {
	tests := []struct {
		input string
		want  []token
	}{
		{input: "asdf", want: []token{
			{TokenType: typeStringValue, Value: "asdf"},
		}},
		{input: "-s", want: []token{
			{TokenType: typeShortGroup, Value: "-s"},
		}},
		{input: "--this-is-long-key", want: []token{
			{TokenType: typeLongKey, Value: "--this-is-long-key"},
		}},
		{input: "-s --this-is-long-key", want: []token{
			{TokenType: typeShortGroup, Value: "-s"},
			{TokenType: typeLongKey, Value: "--this-is-long-key"},
		}},
		{
			input: `--key "spaced value"`,
			want: []token{
				{TokenType: typeLongKey, Value: "--key"},
				{TokenType: typeStringValue, Value: "spaced value"},
			},
		},
		{
			input: `-abc --key "\\spaced \"value\""`,
			want: []token{
				{TokenType: typeShortGroup, Value: "-abc"},
				{TokenType: typeLongKey, Value: "--key"},
				{TokenType: typeStringValue, Value: `\spaced "value"`},
			},
		},
		{
			input: `-abc --key "strange value`,
			want: []token{
				{TokenType: typeShortGroup, Value: "-abc"},
				{TokenType: typeLongKey, Value: "--key"},
				{TokenType: typeStringValue, Value: `strange value`},
			},
		},
		// TODO: the case above with unclosed quote should most likely be an error
		{
			input: `-abc --key "strange value"" -def`,
			want: []token{
				{TokenType: typeShortGroup, Value: "-abc"},
				{TokenType: typeLongKey, Value: "--key"},
				{TokenType: typeStringValue, Value: `strange value`},
				{TokenType: typeStringValue, Value: ` -def`},
			},
		},
		{
			input: `"a""b""c"`,
			want: []token{
				{TokenType: typeStringValue, Value: `a`},
				{TokenType: typeStringValue, Value: `b`},
				{TokenType: typeStringValue, Value: `c`},
			},
		},
		{
			input: `--k@y-" abc`,
			want: []token{
				{TokenType: typeLongKey, Value: `--k@y-"`},
				{TokenType: typeStringValue, Value: `abc`},
			},
		},
	}

	for _, test := range tests {
		got := lex(test.input)
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("lex(%q) = %v, want %v", test.input, got, test.want)
		}
	}
}
