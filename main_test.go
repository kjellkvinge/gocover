package main

import (
	"testing"

	"golang.org/x/tools/cover"
)

func TestProfileSrc(t *testing.T) {
	profiles, err := cover.ParseProfiles("testdata/coverage.out")
	if err != nil {
		t.Error(err)
	}
	fFileName = "foo.go"
	if percentCovered(profiles[0]) != 87.5 {
		t.Error("Exected 87.5% coverage")
	}
	printFile(profiles, fFileName)
}

func TestProfileFunc(t *testing.T) {
	profiles, err := cover.ParseProfiles("testdata/coverage.out")
	if err != nil {
		t.Error(err)
	}
	fFunc = "foo"
	if percentCovered(profiles[0]) != 87.5 {
		t.Error("Exected 87.5% coverage")
	}
	printfunc("foo", profiles)
}

func TestReport(t *testing.T) {
	profiles, err := cover.ParseProfiles("testdata/coverage.out")
	if err != nil {
		t.Error(err)
	}
	generateReport(profiles)
}
