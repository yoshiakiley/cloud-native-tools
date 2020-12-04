package utils

import (
	"io/ioutil"
	"reflect"
	"strings"
	"testing"
)

func Test_listDirectory(t *testing.T) {
	var files = []string{"test.conf"}

	dir, err := ioutil.TempDir(".", "tmp")
	if err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(strings.Join([]string{dir, "test.conf"}, "/"), []byte(""), 0777); err != nil {
		t.Fatal(err)
	}
	expectedFiles, err := ListDirectory(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(expectedFiles, files) {
		t.Fatal("not match expected result")
	}
}
