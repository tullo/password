package password

import (
	"io"
	"strings"
	"sync/atomic"
	"testing"
)

type (
	MockReader struct {
		Counter int64
	}
)

const (
	N = 10000
)

func (mr *MockReader) Read(data []byte) (int, error) {
	for i := 0; i < len(data); i++ {
		data[i] = byte(atomic.AddInt64(&mr.Counter, 1))
	}
	return len(data), nil
}

func testHasDuplicates(tb testing.TB, s string) bool {
	found := make(map[rune]struct{}, len(s))
	for _, ch := range s {
		if _, ok := found[ch]; ok {
			return true
		}
		found[ch] = struct{}{}
	}
	return false
}

func testGeneratorGenerate(t *testing.T, reader io.Reader) {
	t.Parallel()

	gen, err := NewGenerator(nil)
	if reader != nil {
		gen.reader = reader
	}
	if err != nil {
		t.Fatal(err)
	}

	t.Run("exceeds_length", func(t *testing.T) {
		t.Parallel()

		if _, err := gen.Generate(0, 1, 0, true, false); err != ErrExceedsTotalLength {
			t.Errorf("expected %q to be %q", err, ErrExceedsTotalLength)
		}

		if _, err := gen.Generate(0, 0, 1, true, false); err != ErrExceedsTotalLength {
			t.Errorf("expected %q to be %q", err, ErrExceedsTotalLength)
		}
	})

	t.Run("exceeds_letters_available", func(t *testing.T) {
		t.Parallel()

		if _, err := gen.Generate(1000, 0, 0, true, false); err != ErrLettersExceedsAvailable {
			t.Errorf("expected %q to be %q", err, ErrLettersExceedsAvailable)
		}
	})

	t.Run("exceeds_digits_available", func(t *testing.T) {
		t.Parallel()

		if _, err := gen.Generate(52, 11, 0, true, false); err != ErrDigitsExceedsAvailable {
			t.Errorf("expected %q to be %q", err, ErrDigitsExceedsAvailable)
		}
	})

	t.Run("exceeds_symbols_available", func(t *testing.T) {
		t.Parallel()

		if _, err := gen.Generate(52, 0, 31, true, false); err != ErrSymbolsExceedsAvailable {
			t.Errorf("expected %q to be %q", err, ErrSymbolsExceedsAvailable)
		}
	})

	t.Run("gen_lowercase", func(t *testing.T) {
		t.Parallel()

		for i := 0; i < N; i++ {
			res, err := gen.Generate(i%len(LowerLetters), 0, 0, false, true)
			if err != nil {
				t.Error(err)
			}

			if res != strings.ToLower(res) {
				t.Errorf("%q is not lowercase", res)
			}
		}
	})

	t.Run("gen_uppercase", func(t *testing.T) {
		t.Parallel()

		res, err := gen.Generate(1000, 0, 0, true, true)
		if err != nil {
			t.Error(err)
		}

		if res == strings.ToLower(res) {
			t.Errorf("%q does not include uppercase", res)
		}
	})

	t.Run("gen_no_repeats", func(t *testing.T) {
		t.Parallel()

		for i := 0; i < N; i++ {
			res, err := gen.Generate(52, 10, 30, false, false)
			if err != nil {
				t.Error(err)
			}

			if testHasDuplicates(t, res) {
				t.Errorf("%q should not have duplicates", res)
			}
		}
	})
}

func TestGeneratorGenerate(t *testing.T) {
	testGeneratorGenerate(t, nil)
}

func TestGenerator_Reader_Generate(t *testing.T) {
	testGeneratorGenerate(t, &MockReader{})
}

func testGeneratorGenerateCustom(t *testing.T, reader io.Reader) {
	t.Parallel()

	gen, err := NewGenerator(&GeneratorInput{
		LowerLetters: "abcde",
		UpperLetters: "ABCDE",
		Symbols:      "!@#$%",
		Digits:       "01234",
		Reader:       reader,
	})
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < N; i++ {
		res, err := gen.Generate(52, 10, 10, false, true)
		if err != nil {
			t.Error(err)
		}

		if strings.Contains(res, "f") {
			t.Errorf("%q should only contain lower letters abcde", res)
		}

		if strings.Contains(res, "F") {
			t.Errorf("%q should only contain upper letters ABCDE", res)
		}

		if strings.Contains(res, "&") {
			t.Errorf("%q should only include symbols !@#$%%", res)
		}

		if strings.Contains(res, "5") {
			t.Errorf("%q should only contain digits 01234", res)
		}
	}
}

func TestGeneratorGenerateCustom(t *testing.T) {
	testGeneratorGenerateCustom(t, nil)
}

func TestGenerator_Reader_Generate_Custom(t *testing.T) {
	testGeneratorGenerateCustom(t, &MockReader{})
}

func Test_containsUpper(t *testing.T) {

	var TestCases = []struct {
		Name           string
		InputString    string
		ExpectedOutput bool
	}{
		{
			Name:           "Has Upper",
			InputString:    "Maryhadalittlelamb",
			ExpectedOutput: true,
		},
		{
			Name:           "No Upper",
			InputString:    "maryhadalittlelamb",
			ExpectedOutput: false,
		},
	}

	for _, test := range TestCases {
		res := containsUpper(test.InputString)

		if res != test.ExpectedOutput {
			t.Errorf("Testcase %s failed. want - %t, got - %t", test.Name, test.ExpectedOutput, res)
		}
	}
}

func Test_containsLower(t *testing.T) {

	var TestCases = []struct {
		Name           string
		InputString    string
		ExpectedOutput bool
	}{
		{
			Name:           "All Upper",
			InputString:    "MARYHADALITTLELAMB",
			ExpectedOutput: false,
		},
		{
			Name:           "All Lower",
			InputString:    "maryhadalittlelamb",
			ExpectedOutput: true,
		},
	}

	for _, test := range TestCases {
		res := containsLower(test.InputString)

		if res != test.ExpectedOutput {
			t.Errorf("Testcase %s failed. want - %t, got - %t", test.Name, test.ExpectedOutput, res)
		}
	}
}

func Test_containsDigits(t *testing.T) {

	var TestCases = []struct {
		Name           string
		InputString    string
		ExpectedOutput bool
	}{
		{
			Name:           "No Digit",
			InputString:    "MARYHADALITTLELAMB",
			ExpectedOutput: false,
		},
		{
			Name:           "Has Digit",
			InputString:    "maryhadalittlelamb1",
			ExpectedOutput: true,
		},
	}

	for _, test := range TestCases {
		res := containsDigit(test.InputString)

		if res != test.ExpectedOutput {
			t.Errorf("Testcase %s failed. want - %t, got - %t", test.Name, test.ExpectedOutput, res)
		}
	}
}

func Test_containsSymbol(t *testing.T) {

	var TestCases = []struct {
		Name           string
		InputString    string
		ExpectedOutput bool
	}{
		{
			Name:           "No Symbol",
			InputString:    "MARYHADALITTLELAMB",
			ExpectedOutput: false,
		},
		{
			Name:           "Has Symbol",
			InputString:    "mary}hadalittlelamb",
			ExpectedOutput: true,
		},
		{
			Name:           "empty",
			InputString:    "",
			ExpectedOutput: false,
		},
	}

	for _, tc := range TestCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			res := containsSymbol(tc.InputString)

			if res != tc.ExpectedOutput {
				t.Errorf("Testcase %s failed. want - %t, got - %t", tc.Name, tc.ExpectedOutput, res)
			}
		})
	}
}
func Test_isLegalPassword(t *testing.T) {

	var TestCases = []struct {
		Name           string
		Password       string
		NeedsLower     bool
		NeedsUpper     bool
		NeedsDigit     bool
		NeedsSymbol    bool
		ExpectedOutput bool
	}{
		{
			Name:           "Needs everything, only upper",
			Password:       "MARYHADALITTLELAMB",
			NeedsLower:     true,
			NeedsUpper:     true,
			NeedsDigit:     true,
			NeedsSymbol:    true,
			ExpectedOutput: false,
		},
		{
			Name:           "Needs everything, only upper & sumbol",
			Password:       "MARYHADALITTLELAMB$",
			NeedsLower:     true,
			NeedsUpper:     true,
			NeedsDigit:     true,
			NeedsSymbol:    true,
			ExpectedOutput: false,
		},
		{
			Name:           "Needs everything, missing lower",
			Password:       "M4RYHADALITTLELAMB$",
			NeedsLower:     true,
			NeedsUpper:     true,
			NeedsDigit:     true,
			NeedsSymbol:    true,
			ExpectedOutput: false,
		},
		{
			Name:           "Needs everything, has everything",
			Password:       "M4RYHADAlittleLAMB$",
			NeedsLower:     true,
			NeedsUpper:     true,
			NeedsDigit:     true,
			NeedsSymbol:    true,
			ExpectedOutput: true,
		},
		{
			Name:           "Needs lower has lower",
			Password:       "maryhadalittlelamb",
			NeedsLower:     true,
			NeedsUpper:     false,
			NeedsDigit:     false,
			NeedsSymbol:    false,
			ExpectedOutput: true,
		},
		{
			Name:           "Needs lower, has upper",
			Password:       "MARYHADALITTLELAMB",
			NeedsLower:     true,
			NeedsUpper:     false,
			NeedsDigit:     false,
			NeedsSymbol:    false,
			ExpectedOutput: false,
		},
		{
			Name:           "Needs upper, has upper",
			Password:       "MARYHADALITTLELAMB",
			NeedsLower:     false,
			NeedsUpper:     true,
			NeedsDigit:     false,
			NeedsSymbol:    false,
			ExpectedOutput: true,
		},
		{
			Name:           "Needs upper, has lower",
			Password:       "maryhadalittlelamb",
			NeedsLower:     false,
			NeedsUpper:     true,
			NeedsDigit:     false,
			NeedsSymbol:    false,
			ExpectedOutput: false,
		},
		{
			Name:           "Needs digit, has digit",
			Password:       "M4RYHADALITTLELAMB",
			NeedsLower:     false,
			NeedsUpper:     false,
			NeedsDigit:     true,
			NeedsSymbol:    false,
			ExpectedOutput: true,
		},
		{
			Name:           "Needs digit, has no digit",
			Password:       "maryhadalittlelamb",
			NeedsLower:     false,
			NeedsUpper:     false,
			NeedsDigit:     true,
			NeedsSymbol:    false,
			ExpectedOutput: false,
		},
		{
			Name:           "Needs symbol, has symbol",
			Password:       "MaRYHADALITTLELAM&",
			NeedsLower:     false,
			NeedsUpper:     false,
			NeedsDigit:     false,
			NeedsSymbol:    true,
			ExpectedOutput: true,
		},
		{
			Name:           "Needs symbol, has no symbol",
			Password:       "maryhadalittlelamb",
			NeedsLower:     false,
			NeedsUpper:     false,
			NeedsDigit:     false,
			NeedsSymbol:    true,
			ExpectedOutput: false,
		},
	}

	for _, tc := range TestCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			res := isLegalPassword(tc.Password, tc.NeedsLower, tc.NeedsUpper, tc.NeedsDigit, tc.NeedsSymbol)

			if res != tc.ExpectedOutput {
				t.Errorf("Testcase %s failed. want - %t, got - %t", tc.Name, tc.ExpectedOutput, res)
			}
		})
	}
}
