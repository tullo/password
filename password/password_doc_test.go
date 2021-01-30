package password_test

import (
	"fmt"
	"log"

	"github.com/tullo/password/password"
)

func ExampleGenerate() {
	res, err := password.Generate(64, 10, 10, false, false)
	if err != nil {
		log.Fatal(err)
	}
	log.Print(res)
}

func ExampleMustGenerate() {
	// Will panic on error
	res := password.MustGenerate(64, 10, 10, false, false)
	log.Print(res)
}

func ExampleStatefulGenerator_Generate() {
	gen, err := password.NewStatefulGenerator(nil)
	if err != nil {
		log.Fatal(err)
	}

	res, err := gen.Generate(64, 10, 10, false, false)
	if err != nil {
		log.Fatal(err)
	}
	log.Print(res)
}

func ExampleNewStatefulGenerator_nil() {
	// This is exactly the same as calling "Generate" directly.
	// It will use all the default values.
	gen, err := password.NewStatefulGenerator(nil)
	if err != nil {
		log.Fatal(err)
	}

	_ = gen // gen.Generate(...)
}

func ExampleNewStatefulGenerator_custom() {
	// Customize the list of symbols.
	gen, err := password.NewStatefulGenerator(&password.GeneratorInput{
		Symbols: "!@#$%^()",
	})
	if err != nil {
		log.Fatal(err)
	}

	_ = gen // gen.Generate(...)
}

func ExampleNewMockPasswordGenerator_testing() {
	// Accept a password.Generator interface instead of a
	// password.Generator struct.
	f := func(g password.Generator) string {
		// These values don't matter
		return g.MustGenerate(1, 2, 3, false, false)
	}

	// In tests
	gen := password.NewMockPasswordGenerator("canned-response", nil)

	fmt.Print(f(gen))
	// Output: canned-response
}
