//go:build mage

// Tasks for working on the site. Run `go tool mage` to list them.
package main

import (
	"os"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// Serve runs the dev server on localhost:1313 and rebuilds on save.
func Serve() error {
	return sh.RunV("go", "run", "./cmd/wrkstdy", "serve")
}

// Build writes the site to public/. Set BASEURL to override site.toml's baseURL.
func Build() error {
	args := []string{"run", "./cmd/wrkstdy", "build"}
	if url := os.Getenv("BASEURL"); url != "" {
		args = append(args, "-baseurl", url)
	}
	return sh.RunV("go", args...)
}

// Test runs go vet and the tests.
func Test() error {
	if err := sh.RunV("go", "vet", "./..."); err != nil {
		return err
	}
	return sh.RunV("go", "test", "./...")
}

// Check runs the tests, then builds. Netlify runs this on every deploy.
func Check() {
	mg.SerialDeps(Test, Build)
}
