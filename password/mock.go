package password

// Built-time checks that the generators implement the interface.
var _ Generator = (*MockPasswordGenerator)(nil)

// MockPasswordGenerator is a generator that satisfies the Generator interface.
type MockPasswordGenerator struct {
	result string
	err    error
}

// NewMockPasswordGenerator creates a new mock generator. If an error is
// provided, the error is returned. If a result if provided, the result is
// always returned, regardless of what parameters are passed into the Generate
// or MustGenerate methods.
//
// This function is most useful for tests where you want to have predicable
// results for a transitive resource that depends on the password package.
func NewMockPasswordGenerator(result string, err error) *MockPasswordGenerator {
	return &MockPasswordGenerator{
		result: result,
		err:    err,
	}
}

// Generate returns the mocked result or error.
func (g *MockPasswordGenerator) Generate(int, int, int, bool, bool) (string, error) {
	if g.err != nil {
		return "", g.err
	}
	return g.result, nil
}

// GenerateWithPolicy returns the mocked result or error.
func (g *MockPasswordGenerator) GenerateWithPolicy(int, int, int, bool, bool, bool, bool, bool, bool) (string, error) {
	if g.err != nil {
		return "", g.err
	}
	return g.result, nil
}

// MustGenerate returns the mocked result or panics if an error was given.
func (g *MockPasswordGenerator) MustGenerate(int, int, int, bool, bool) string {
	if g.err != nil {
		panic(g.err)
	}
	return g.result
}
