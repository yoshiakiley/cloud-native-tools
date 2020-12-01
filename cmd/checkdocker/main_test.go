package main

import (
	"io/ioutil"
	"os"
	"testing"
)

func Test_CheckDockerFile(t *testing.T) {
	dir, err := ioutil.TempDir(".", "tmp")
	if err != nil {
		t.Fatal(err)
	}
	err = CheckDockerFile(dir, "django")
	if err != nil {
		t.Fatal(err)
	}
	err = CheckDockerFile(dir, "django")
	if err != nil {
		t.Fatal(err)
	}

}

func Test_djangoDocker(t *testing.T) {
	err := djangoDocker("./Dockerfile")
	if err != nil {
		t.Fatal(err)
	}
	var exist = true
	if _, err := os.Stat("./Dockerfile"); os.IsNotExist(err) {
		exist = false
	}
	if !exist {
		t.Fatal("not match expected result")
	}

}
