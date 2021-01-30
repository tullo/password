// Package password provides a library for generating high-entropy random
// password strings via the crypto/rand package.
//
//    res, err := Generate(64, 10, 10, false, false)
//    if err != nil  {
//      log.Fatal(err)
//    }
//    log.Printf(res)
//
// Most functions are safe for concurrent use.
package password

import (
	"crypto/rand"
	"errors"
	"io"
	"math/big"
	"regexp"
	"strings"
)

// Built-time checks that the generators implement the interface.
var _ Generator = (*StatefulGenerator)(nil)

// Generator is an interface that implements the Generate function. This
// is useful for testing where you can pass this interface instead of a real
// password generator to mock responses for predicability.
type Generator interface {
	Generate(int, int, int, bool, bool) (string, error)
	MustGenerate(int, int, int, bool, bool) string
	GenerateWithPolicy(int, int, int, bool, bool, bool, bool, bool, bool) (string, error)
}

const (
	// LowerLetters is the list of lowercase letters.
	LowerLetters = "abcdefghijklmnopqrstuvwxyz"

	// UpperLetters is the list of uppercase letters.
	UpperLetters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

	// Digits is the list of permitted digits.
	Digits = "0123456789"

	// Symbols is the list of symbols.
	Symbols = "~!@#$%^&*()_+`-={}|[]\\:\"<>?,./"
)

var (
	// ErrExceedsTotalLength is the error returned with the number of digits and
	// symbols is greater than the total length.
	ErrExceedsTotalLength = errors.New("number of digits and symbols must be less than total length")

	// ErrLettersExceedsAvailable is the error returned with the number of letters
	// exceeds the number of available letters and repeats are not allowed.
	ErrLettersExceedsAvailable = errors.New("number of letters exceeds available letters and repeats are not allowed")

	// ErrDigitsExceedsAvailable is the error returned with the number of digits
	// exceeds the number of available digits and repeats are not allowed.
	ErrDigitsExceedsAvailable = errors.New("number of digits exceeds available digits and repeats are not allowed")

	// ErrSymbolsExceedsAvailable is the error returned with the number of symbols
	// exceeds the number of available symbols and repeats are not allowed.
	ErrSymbolsExceedsAvailable = errors.New("number of symbols exceeds available symbols and repeats are not allowed")
)

// StatefulGenerator is a generator which can be used to customize the list
// of letters, digits, and/or symbols.
type StatefulGenerator struct {
	lowerLetters string
	upperLetters string
	digits       string
	symbols      string
	reader       io.Reader
}

// GeneratorInput is used as input to the NewStatefulGenerator function.
type GeneratorInput struct {
	LowerLetters string
	UpperLetters string
	Digits       string
	Symbols      string
	Reader       io.Reader // rand.Reader by default
}

// NewStatefulGenerator creates a new StatefulGenerator from the specified
// configuration. If no input is given, all the default values are used. This
// function is safe for concurrent use.
func NewStatefulGenerator(i *GeneratorInput) (*StatefulGenerator, error) {
	if i == nil {
		i = new(GeneratorInput)
	}

	g := &StatefulGenerator{
		lowerLetters: i.LowerLetters,
		upperLetters: i.UpperLetters,
		digits:       i.Digits,
		symbols:      i.Symbols,
		reader:       i.Reader,
	}

	if g.lowerLetters == "" {
		g.lowerLetters = LowerLetters
	}

	if g.upperLetters == "" {
		g.upperLetters = UpperLetters
	}

	if g.digits == "" {
		g.digits = Digits
	}

	if g.symbols == "" {
		g.symbols = Symbols
	}

	if g.reader == nil {
		g.reader = rand.Reader
	}

	return g, nil
}

// Generate generates a password with the given requirements. length is the
// total number of characters in the password. numDigits is the number of digits
// to include in the result. numSymbols is the number of symbols to include in
// the result. noUpper excludes uppercase letters from the results. allowRepeat
// allows characters to repeat.
//
// The algorithm is fast, but it's not designed to be performant; it favors
// entropy over speed. This function is safe for concurrent use.
func (g *StatefulGenerator) Generate(length, numDigits, numSymbols int, includeUpper, allowRepeat bool) (string, error) {
	letters := g.lowerLetters
	if includeUpper {
		letters += g.upperLetters
	}

	chars := length - numDigits - numSymbols
	if chars < 0 {
		return "", ErrExceedsTotalLength
	}

	if !allowRepeat && chars > len(letters) {
		return "", ErrLettersExceedsAvailable
	}

	if !allowRepeat && numDigits > len(g.digits) {
		return "", ErrDigitsExceedsAvailable
	}

	if !allowRepeat && numSymbols > len(g.symbols) {
		return "", ErrSymbolsExceedsAvailable
	}

	var result string

	// Characters
	for i := 0; i < chars; i++ {
		ch, err := randomElement(g.reader, letters)
		if err != nil {
			return "", err
		}

		if !allowRepeat && strings.Contains(result, ch) {
			i--
			continue
		}

		result, err = randomInsert(g.reader, result, ch)
		if err != nil {
			return "", err
		}
	}

	// Digits
	for i := 0; i < numDigits; i++ {
		d, err := randomElement(g.reader, g.digits)
		if err != nil {
			return "", err
		}

		if !allowRepeat && strings.Contains(result, d) {
			i--
			continue
		}

		result, err = randomInsert(g.reader, result, d)
		if err != nil {
			return "", err
		}
	}

	// Symbols
	for i := 0; i < numSymbols; i++ {
		sym, err := randomElement(g.reader, g.symbols)
		if err != nil {
			return "", err
		}

		if !allowRepeat && strings.Contains(result, sym) {
			i--
			continue
		}

		result, err = randomInsert(g.reader, result, sym)
		if err != nil {
			return "", err
		}
	}

	return result, nil
}

// MustGenerate is the same as Generate, but panics on error.
func (g *StatefulGenerator) MustGenerate(length, numDigits, numSymbols int, includeUpper, allowRepeat bool) string {
	res, err := g.Generate(length, numDigits, numSymbols, includeUpper, allowRepeat)
	if err != nil {
		panic(err)
	}
	return res
}

// GenerateWithPolicy is the same as Generate, but ensures result matches specified policy
func (g *StatefulGenerator) GenerateWithPolicy(length, numDigits, numSymbols int, includeUpper, allowRepeat, needsLower, needsUpper, needsDigit, needsSymbol bool) (result string, err error) {

	for {
		result, err = g.Generate(length, numDigits, numSymbols, includeUpper, allowRepeat)
		if err != nil {
			return "", err
		}
		if isLegalPassword(result, needsLower, needsUpper, needsDigit, needsSymbol) {
			break
		}
	}

	return result, nil
}

// Generate is the package shortcut for Generator.Generate.
func Generate(length, numDigits, numSymbols int, includeUpper, allowRepeat bool) (string, error) {
	gen, err := NewStatefulGenerator(nil)
	if err != nil {
		return "", err
	}

	return gen.Generate(length, numDigits, numSymbols, includeUpper, allowRepeat)
}

// GenerateWithPolicy is the package shortcut for Generator.GenerateWithPolicy.
func GenerateWithPolicy(length, numDigits, numSymbols int, includeUpper, allowRepeat, needsLower, needsUpper, needsDigit, needsSymbol bool) (string, error) {
	gen, err := NewStatefulGenerator(nil)
	if err != nil {
		return "", err
	}

	return gen.GenerateWithPolicy(length, numDigits, numSymbols, includeUpper, allowRepeat, needsLower, needsUpper, needsDigit, needsSymbol)
}

// MustGenerate is the package shortcut for Generator.MustGenerate.
func MustGenerate(length, numDigits, numSymbols int, includeUpper, allowRepeat bool) string {
	res, err := Generate(length, numDigits, numSymbols, includeUpper, allowRepeat)
	if err != nil {
		panic(err)
	}
	return res
}

// randomInsert randomly inserts the given value into the given string.
func randomInsert(reader io.Reader, s, val string) (string, error) {
	if s == "" {
		return val, nil
	}

	n, err := rand.Int(reader, big.NewInt(int64(len(s)+1)))
	if err != nil {
		return "", err
	}
	i := n.Int64()
	return s[0:i] + val + s[i:], nil
}

// randomElement extracts a random element from the given string.
func randomElement(reader io.Reader, s string) (string, error) {
	n, err := rand.Int(reader, big.NewInt(int64(len(s))))
	if err != nil {
		return "", err
	}
	return string(s[n.Int64()]), nil
}

func isLegalPassword(p string, needsLower bool, needsUpper bool, needsDigit bool, needsSymbol bool) bool {

	if needsLower && !containsLower(p) {
		return false
	}

	if needsUpper && !containsUpper(p) {
		return false
	}

	if needsDigit && !containsDigit(p) {
		return false
	}

	if needsSymbol && !containsSymbol(p) {
		return false
	}

	return true
}

func containsLower(s string) bool {
	r := regexp.MustCompile(".*[[:lower:]].*")
	return r.MatchString(s)
}

func containsUpper(s string) bool {
	r := regexp.MustCompile(".*[[:upper:]].*")
	return r.MatchString(s)
}

func containsDigit(s string) bool {
	r := regexp.MustCompile(".*[[:digit:]].*")
	return r.MatchString(s)
}

func containsSymbol(s string) bool {
	r := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	return r.MatchString(s)
}
