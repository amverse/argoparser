package argoparser

import (
	"reflect"
	"strings"
	"testing"
)

type TestCase struct {
	Name              string
	Input             string
	Result            any
	ShouldReturnError bool
}

func TestParser(t *testing.T) {
	tc := []TestCase{
		{
			Name:  "Basic test with long keys",
			Input: "--flag-b --value-string value-string --value-int 1 --value-slice-string value1 --value-slice-string value2 --value-slice-int 1 --value-slice-int 2",
			Result: struct {
				FlagA            bool     `arg:"--flag-a"`
				FlagB            bool     `arg:"--flag-b"`
				ValueString      string   `arg:"--value-string"`
				ValueInt         int      `arg:"--value-int"`
				ValueSliceString []string `arg:"--value-slice-string"`
				ValueSliceInt    []int    `arg:"--value-slice-int"`
			}{
				FlagA:            false,
				FlagB:            true,
				ValueString:      "value-string",
				ValueInt:         1,
				ValueSliceString: []string{"value1", "value2"},
				ValueSliceInt:    []int{1, 2},
			},
		},
		{
			Name:  "Test with long keys: unpresented field passed",
			Input: "--flag-bd --value-string value-string --value-int 1 --value-slice-string value1 --value-slice-string value2 --value-slice-int 1 --value-slice-int 2",
			Result: struct {
				FlagA            bool     `arg:"--flag-a"`
				FlagB            bool     `arg:"--flag-b"`
				ValueString      string   `arg:"--value-string"`
				ValueInt         int      `arg:"--value-int"`
				ValueSliceString []string `arg:"--value-slice-string"`
				ValueSliceInt    []int    `arg:"--value-slice-int"`
			}{
				FlagA:            false,
				FlagB:            true,
				ValueString:      "value-string",
				ValueInt:         1,
				ValueSliceString: []string{"value1", "value2"},
				ValueSliceInt:    []int{1, 2},
			},
			ShouldReturnError: true,
		},
		{
			Name:  "Test with long keys: string with spaces",
			Input: "--value-a 'value a' --value-b `value b` --value-c \"value c\"",
			Result: struct {
				ValueA string `arg:"--value-a"`
				ValueB string `arg:"--value-b"`
				ValueC string `arg:"--value-c"`
			}{
				ValueA: "value a",
				ValueB: "value b",
				ValueC: "value c",
			},
		},
		{
			Name:  "Test short flags and groups",
			Input: "-a -cde -f 1 -g -5 -h v1 -h v2 -h v3",
			Result: struct {
				FlagA  bool     `arg:"-a"`
				FlagB  bool     `arg:"-b"`
				FlagC  bool     `arg:"-c"`
				FlagD  bool     `arg:"-d"`
				FlagE  bool     `arg:"-e"`
				ValueF string   `arg:"-f"`
				ValueG int      `arg:"-g"`
				ValueH []string `arg:"-h"`
			}{
				FlagA:  true,
				FlagB:  false,
				FlagC:  true,
				FlagD:  true,
				FlagE:  true,
				ValueF: "1",
				ValueG: -5,
				ValueH: []string{"v1", "v2", "v3"},
			},
		},
		{
			Name:  "Test building index for wrong structure: multiple fields for one key",
			Input: "-f asdf",
			Result: struct {
				Foo string `arg:"-f"`
				Bar string `arg:"-f"`
			}{},
			ShouldReturnError: true,
		},
		{
			Name:  "Test building index for wrong structure: positional field with short or long name",
			Input: "-f asdf",
			Result: struct {
				Foo string `arg:"-f,positional"`
			}{},
			ShouldReturnError: true,
		},
		{
			Name:  "Basic test with positional arguments",
			Input: "value-a 2 value-c",
			Result: struct {
				ValueA string `arg:"positional"`
				ValueB int    `arg:"positional"`
				ValueC string `arg:"positional"`
			}{
				ValueA: "value-a",
				ValueB: 2,
				ValueC: "value-c",
			},
		},
		{
			Name:  "Test with positional arguments: less passed than expected",
			Input: "value-a 2",
			Result: struct {
				ValueA string `arg:"positional"`
				ValueB int    `arg:"positional"`
				ValueC string `arg:"positional"`
			}{
				ValueA: "value-a",
				ValueB: 2,
				ValueC: "",
			},
		},
		{
			Name:  "Test positionals default",
			Input: "value-a 2 a 3 c",
			Result: struct {
				ValueA  string   `arg:"positional"`
				ValueB  int      `arg:"positional"`
				Default []string `arg:"positional"`
			}{
				ValueA:  "value-a",
				ValueB:  2,
				Default: []string{"a", "3", "c"},
			},
		},
		{
			Name:  "Test combined short and long options for one field",
			Input: "--some-value 1 --some-value 2 -v 3",
			Result: struct {
				SomeValue []string `arg:"--some-value,-v"`
			}{
				SomeValue: []string{"1", "2", "3"},
			},
		},
		{
			Name:  "Test required fields: ok",
			Input: "-s sv --long lv pv",
			Result: struct {
				ShortField string `arg:"-s,required"`
				LongField  string `arg:"--long, required"`
				Positional string `arg:"positional,required"`
			}{
				ShortField: "sv",
				LongField:  "lv",
				Positional: "pv",
			},
		},
		{
			Name:  "Test required fields: no short",
			Input: "--long lv pv",
			Result: struct {
				ShortField string `arg:"-s,required"`
				LongField  string `arg:"--long, required"`
				Positional string `arg:"positional,required"`
			}{
				ShortField: "sv",
				LongField:  "lv",
				Positional: "pv",
			},
			ShouldReturnError: true,
		},
		{
			Name:  "Test required fields: no long",
			Input: "-s sv pv",
			Result: struct {
				ShortField string `arg:"-s,required"`
				LongField  string `arg:"--long, required"`
				Positional string `arg:"positional,required"`
			}{
				ShortField: "sv",
				LongField:  "lv",
				Positional: "pv",
			},
			ShouldReturnError: true,
		},
		{
			Name:  "Test required fields: no positional",
			Input: "-s sv --long lv",
			Result: struct {
				ShortField string `arg:"-s,required"`
				LongField  string `arg:"--long, required"`
				Positional string `arg:"positional,required"`
			}{
				ShortField: "sv",
				LongField:  "lv",
				Positional: "pv",
			},
			ShouldReturnError: true,
		},
		{
			Name:  "Weird keys",
			Input: `--" a --@r6ument b --- c`,
			Result: struct {
				FieldA string `arg:"--\""`
				FieldB string `arg:"--@r6ument"`
				FieldC string `arg:"---"`
			}{
				FieldA: "a",
				FieldB: "b",
				FieldC: "c",
			},
		},
		{
			Name:  "Unicode",
			Input: "--long-🔑 🚌 -🍏🍎 -🤠 🤡",
			Result: struct {
				FieldA     string `arg:"--long-🔑"`
				GreenApple bool   `arg:"-🍏"`
				RedApple   bool   `arg:"-🍎"`
				FieldC     bool   `arg:"-🤠"`
				Clown      string `arg:"positional"`
			}{
				FieldA:     "🚌",
				GreenApple: true,
				RedApple:   true,
				FieldC:     true,
				Clown:      "🤡",
			},
		},
		{
			Name:  "Test empty input",
			Input: "",
			Result: struct {
				Default []string `arg:"positional"`
			}{},
		},
		{
			Name:  "Test whitespace only input",
			Input: "   \t\n  ",
			Result: struct {
				Default []string `arg:"positional"`
			}{},
		},
		{
			Name:  "Test multiple positional default fields error",
			Input: "value1 value2",
			Result: struct {
				Default1 []string `arg:"positional"`
				Default2 []string `arg:"positional"`
			}{},
			ShouldReturnError: true,
		},
		{
			Name:  "Test unsupported type error",
			Input: "--value 123",
			Result: struct {
				Value float64 `arg:"--value"`
			}{},
			ShouldReturnError: true,
		},
		{
			Name:  "Test invalid int value error",
			Input: "--value abc",
			Result: struct {
				Value int `arg:"--value"`
			}{},
			ShouldReturnError: true,
		},
		{
			Name:  "Test missing value for non-flag long key",
			Input: "--value",
			Result: struct {
				Value string `arg:"--value"`
			}{},
			ShouldReturnError: true,
		},
		{
			Name:  "Test missing value for non-flag short key",
			Input: "-v",
			Result: struct {
				Value string `arg:"-v"`
			}{},
			ShouldReturnError: true,
		},
		{
			Name:  "Test short group with non-flag field error",
			Input: "-abc",
			Result: struct {
				FlagA bool   `arg:"-a"`
				FlagB bool   `arg:"-b"`
				Value string `arg:"-c"`
			}{},
			ShouldReturnError: true,
		},
		{
			Name:  "Test unknown long key error",
			Input: "--unknown-key value",
			Result: struct {
				Value string `arg:"--value"`
			}{},
			ShouldReturnError: true,
		},
		{
			Name:  "Test unknown short key error",
			Input: "-x value",
			Result: struct {
				Value string `arg:"-v"`
			}{},
			ShouldReturnError: true,
		},
		{
			Name:  "Test unexpected positional parameter error",
			Input: "value1 value2",
			Result: struct {
				Value string `arg:"--value"`
			}{},
			ShouldReturnError: true,
		},
		{
			Name:  "Test empty string values",
			Input: "--value1 '' --value2 \"\" --value3 ``",
			Result: struct {
				Value1 string `arg:"--value1"`
				Value2 string `arg:"--value2"`
				Value3 string `arg:"--value3"`
			}{
				Value1: "",
				Value2: "",
				Value3: "",
			},
		},
		{
			Name:  "Test special characters in values",
			Input: `--value1 "a\nb" --value2 "a\tb" --value3 "a\\b"`,
			Result: struct {
				Value1 string `arg:"--value1"`
				Value2 string `arg:"--value2"`
				Value3 string `arg:"--value3"`
			}{
				Value1: "a\nb",
				Value2: "a\tb",
				Value3: "a\\b",
			},
		},
		{
			Name:  "Test escaped special characters in values",
			Input: `--value1 "a\\nb" --value2 "a\\tb" --value3 "a\\b"`,
			Result: struct {
				Value1 string `arg:"--value1"`
				Value2 string `arg:"--value2"`
				Value3 string `arg:"--value3"`
			}{
				Value1: `a\nb`,
				Value2: `a\tb`,
				Value3: `a\b`,
			},
		},
		{
			Name:  "Test very long values",
			Input: "--long-value " + strings.Repeat("a", 1000),
			Result: struct {
				LongValue string `arg:"--long-value"`
			}{
				LongValue: strings.Repeat("a", 1000),
			},
		},
		{
			Name:  "Test large slice values",
			Input: strings.Repeat("--value x ", 100),
			Result: struct {
				Values []string `arg:"--value"`
			}{
				Values: strings.Split(strings.TrimSpace(strings.Repeat("x ", 100)), " "),
			},
		},
		{
			Name:  "Test mixed positional and named arguments",
			Input: "pos1 --flag1 --value1 val1 pos2 --flag2 pos3",
			Result: struct {
				Pos1   string `arg:"positional"`
				Pos2   string `arg:"positional"`
				Pos3   string `arg:"positional"`
				Flag1  bool   `arg:"--flag1"`
				Flag2  bool   `arg:"--flag2"`
				Value1 string `arg:"--value1"`
			}{
				Pos1:   "pos1",
				Pos2:   "pos2",
				Pos3:   "pos3",
				Flag1:  true,
				Flag2:  true,
				Value1: "val1",
			},
		},
		{
			Name:  "Test duplicate keys in different fields error",
			Input: "--value x",
			Result: struct {
				Value1 string `arg:"--value"`
				Value2 string `arg:"--value"`
			}{},
			ShouldReturnError: true,
		},
		{
			Name:  "Test invalid tag format error",
			Input: "--value x",
			Result: struct {
				Value string `arg:"invalid-tag"`
			}{},
			ShouldReturnError: true,
		},
		{
			Name:  "Test positional field with short name error",
			Input: "value",
			Result: struct {
				Value string `arg:"positional,-v"`
			}{},
			ShouldReturnError: true,
		},
		{
			Name:  "Test positional field with long name error",
			Input: "value",
			Result: struct {
				Value string `arg:"positional,--value"`
			}{},
			ShouldReturnError: true,
		},
		{
			Name:  "Test negative int values",
			Input: "--value1 -123 --value2 -0 --value3 -999999",
			Result: struct {
				Value1 int `arg:"--value1"`
				Value2 int `arg:"--value2"`
				Value3 int `arg:"--value3"`
			}{
				Value1: -123,
				Value2: 0,
				Value3: -999999,
			},
		},
		{
			Name:  "Test zero values",
			Input: "--string-value '' --int-value 0",
			Result: struct {
				StringValue string `arg:"--string-value"`
				IntValue    int    `arg:"--int-value"`
			}{
				StringValue: "",
				IntValue:    0,
			},
		},
		{
			Name:  "Test mixed bool and value fields in short group",
			Input: "-abc -d value",
			Result: struct {
				A bool   `arg:"-a"`
				B bool   `arg:"-b"`
				C bool   `arg:"-c"`
				D string `arg:"-d"`
			}{
				A: true, B: true, C: true, D: "value",
			},
		},
		{
			Name:  "Test quoted values with spaces",
			Input: `--value1 "hello world" --value2 'special:chars@#$%'`,
			Result: struct {
				Value1 string `arg:"--value1"`
				Value2 string `arg:"--value2"`
			}{
				Value1: "hello world",
				Value2: "special:chars@#$%",
			},
		},
		{
			Name:  "Test edge case: single dash",
			Input: "-",
			Result: struct {
				Flag bool `arg:"--flag"`
			}{
				Flag: false,
			},
			ShouldReturnError: true,
		},
		{
			Name:  "Test edge case: double dash only",
			Input: "--",
			Result: struct {
				Flag bool `arg:"--flag"`
			}{
				Flag: false,
			},
			ShouldReturnError: true,
		},
		{
			Name:  "Test edge case: mixed quotes in same value",
			Input: `--value "mixed'quotes"`,
			Result: struct {
				Value string `arg:"--value"`
			}{
				Value: "mixed'quotes",
			},
		},
		{
			Name:  "Test edge case: escaped quotes",
			Input: `--value "escaped\"quote"`,
			Result: struct {
				Value string `arg:"--value"`
			}{
				Value: "escaped\"quote",
			},
		},
		{
			Name:  "Test edge case: backtick with spaces",
			Input: "--value `command with spaces`",
			Result: struct {
				Value string `arg:"--value"`
			}{
				Value: "command with spaces",
			},
		},
		{
			Name:  "Test edge case: empty positional arguments",
			Input: "",
			Result: struct {
				Pos1 string `arg:"positional"`
				Pos2 int    `arg:"positional"`
			}{
				Pos1: "",
				Pos2: 0,
			},
		},
		{
			Name:  "Test edge case: positional after named arguments",
			Input: "--flag1 --value1 val1 pos1 pos2",
			Result: struct {
				Flag1  bool   `arg:"--flag1"`
				Value1 string `arg:"--value1"`
				Pos1   string `arg:"positional"`
				Pos2   string `arg:"positional"`
			}{
				Flag1: true, Value1: "val1", Pos1: "pos1", Pos2: "pos2",
			},
		},
		{
			Name:  "Test edge case: multiple spaces between arguments",
			Input: "  --flag1    --value1    val1    pos1    ",
			Result: struct {
				Flag1  bool   `arg:"--flag1"`
				Value1 string `arg:"--value1"`
				Pos1   string `arg:"positional"`
			}{
				Flag1: true, Value1: "val1", Pos1: "pos1",
			},
		},
		{
			Name:  "Test edge case: tab characters in input",
			Input: "--flag1\t--value1\tval1\tpos1",
			Result: struct {
				Flag1  bool   `arg:"--flag1"`
				Value1 string `arg:"--value1"`
				Pos1   string `arg:"positional"`
			}{
				Flag1: true, Value1: "val1", Pos1: "pos1",
			},
		},
		{
			Name:  "Test edge case: newline characters in input",
			Input: "--flag1\n--value1\nval1\npos1",
			Result: struct {
				Flag1  bool   `arg:"--flag1"`
				Value1 string `arg:"--value1"`
				Pos1   string `arg:"positional"`
			}{
				Flag1: true, Value1: "val1", Pos1: "pos1",
			},
		},
	}

	for _, testCase := range tc {
		impl(t, testCase)
	}
}

func TestParseSlice(t *testing.T) {
	t.Run("Test ParseSlice with string array", func(t *testing.T) {
		input := []string{"pos1", "--flag1", "--value1", "val1"}
		result := struct {
			Pos1   string `arg:"positional"`
			Flag1  bool   `arg:"--flag1"`
			Value1 string `arg:"--value1"`
		}{}

		err := ParseSlice(input, &result)
		if err != nil {
			t.Fatalf("ParseSlice failed: %s", err)
		}

		expected := struct {
			Pos1   string `arg:"positional"`
			Flag1  bool   `arg:"--flag1"`
			Value1 string `arg:"--value1"`
		}{
			Pos1: "pos1", Flag1: true, Value1: "val1",
		}

		if !reflect.DeepEqual(result, expected) {
			t.Fatalf("expected %v, got %v", expected, result)
		}
	})

	t.Run("Test ParseSlice with empty array", func(t *testing.T) {
		input := []string{}
		result := struct {
			Flag bool `arg:"--flag"`
		}{}

		err := ParseSlice(input, &result)
		if err != nil {
			t.Fatalf("ParseSlice failed: %s", err)
		}

		expected := struct {
			Flag bool `arg:"--flag"`
		}{
			Flag: false,
		}

		if !reflect.DeepEqual(result, expected) {
			t.Fatalf("expected %v, got %v", expected, result)
		}
	})

	t.Run("Test ParseSlice with single element", func(t *testing.T) {
		input := []string{"--flag"}
		result := struct {
			Flag bool `arg:"--flag"`
		}{}

		err := ParseSlice(input, &result)
		if err != nil {
			t.Fatalf("ParseSlice failed: %s", err)
		}

		expected := struct {
			Flag bool `arg:"--flag"`
		}{
			Flag: true,
		}

		if !reflect.DeepEqual(result, expected) {
			t.Fatalf("expected %v, got %v", expected, result)
		}
	})
}

func TestInputValidation(t *testing.T) {
	t.Run("Test nil pointer error", func(t *testing.T) {
		err := ParseString("--flag", nil)
		if err == nil {
			t.Fatal("expected error for nil pointer")
		}
	})

	t.Run("Test non-pointer error", func(t *testing.T) {
		var result struct {
			Flag bool `arg:"--flag"`
		}
		err := ParseString("--flag", result)
		if err == nil {
			t.Fatal("expected error for non-pointer")
		}
	})

	t.Run("Test non-struct pointer error", func(t *testing.T) {
		var result string
		err := ParseString("--flag", &result)
		if err == nil {
			t.Fatal("expected error for non-struct pointer")
		}
	})
}

func impl(t *testing.T, testCase TestCase) {
	t.Run(testCase.Name, func(t *testing.T) {
		result := reflect.New(reflect.TypeOf(testCase.Result))
		if err := ParseString(testCase.Input, result.Interface()); err != nil {
			if testCase.ShouldReturnError {
				return
			}
			t.Fatalf("failed to parse input: %s", err)
		}
		if !reflect.DeepEqual(result.Elem().Interface(), testCase.Result) {
			t.Fatalf("expected %v, got %v (%T, %T)", testCase.Result, result.Elem().Interface(), testCase.Result, result.Elem().Interface())
		}
	})
}
