package argoparser

import (
	"reflect"
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
	}

	for _, testCase := range tc {
		impl(t, testCase)
	}
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
