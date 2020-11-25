package main

import (
	"reflect"
	"testing"
)

var files = []string{"go.mod", "main.go", "main_test.go"}

func Test_search(t *testing.T) {
	expectedFiles := listDirectory(".")
	if !reflect.DeepEqual(expectedFiles, files) {
		t.Fatal("not match expected result")
	}
}
