package main

import (
	"fmt"
	"go/build"
	"io"
	"os"
	"path/filepath"
)

// fileExists checks if a file exist or not
func fileExists(f string) bool {
	if _, err := os.Stat(f); os.IsNotExist(err) {
		return false
	}
	return true
}

// findFile finds the location of the named file in GOROOT, GOPATH etc.
func findFile(file string) (string, error) {
	dir, file := filepath.Split(file)
	pkg, err := build.Import(dir, ".", build.FindOnly)
	if err != nil {
		return "", fmt.Errorf("can't find %q: %v", file, err)
	}
	return filepath.Join(pkg.Dir, file), nil
}

// print

// legend prints sample data colorized
func legend(w io.Writer) {
	fmt.Fprintf(w,
		"| %s | %s | %s | %s %s %s %s %s %s %s %s %s | %s |\n%s\n",
		printUntracked("untracked"),
		fadeprint("no coverage", 0),
		fadeprint("low coverage", 1),
		fadeprint("*", 10),
		fadeprint("*", 20),
		fadeprint("*", 30),
		fadeprint("*", 40),
		fadeprint("*", 50),
		fadeprint("*", 60),
		fadeprint("*", 70),
		fadeprint("*", 80),
		fadeprint("*", 90),
		fadeprint("high coverage", 100),
		"------------------------------------------------------------------------------",
	)

	if !fDebug {
		return
	}

	for i := 0; i <= 10; i += 2 {
		fmt.Fprintf(w, "%.1f%% %s\n", float64(i)/10, fadeprint("test", float64(i)/10))
	}

	for i := 5; i <= 100; i += 5 {
		fmt.Fprintf(w, "%d%% %s\n", i, fadeprint("test", float64(i)))
	}
}
