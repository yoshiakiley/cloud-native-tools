package utils

import (
	"reflect"
	"testing"
)

var files = []string{"go.mod", "main.go", "main_test.go"}

func Test_listDirectory(t *testing.T) {
	//dirName, err := ioutil.TempDir("", "*.go")
	//if err != nil {
	//	t.Fatal(err)
	//}

	expectedFiles, err := ListDirectory(".")
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(expectedFiles, files) {
		t.Fatal("not match expected result")
	}
}
