package main

import (
	"io/ioutil"
	"os"
	"testing"
)

func Test_CheckDockerFile(t *testing.T) {
	dir, err := ioutil.TempDir(".", "tmp")
	defer os.RemoveAll(dir)
	if err != nil {
		t.Fatal(err)
	}
	err = CheckDockerFile(dir, "django", "")
	if err != nil {
		t.Fatal(err)
	}

}

func Test_djangoDocker(t *testing.T) {
	dockerFile := "./Dockerfile"
	err := djangoDocker(dockerFile)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(dockerFile)
	var exist = true
	if _, err := os.Stat(dockerFile); os.IsNotExist(err) {
		exist = false
	}
	if !exist {
		t.Fatal("not match expected result")
	}
}

func Test_webDocker(t *testing.T) {
	dockerFile := "./Dockerfile"
	err := webDocker(dockerFile)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(dockerFile)
	var exist = true
	if _, err := os.Stat(dockerFile); os.IsNotExist(err) {
		exist = false
	}
	if !exist {
		t.Fatal("not match expected result")
	}
}

func Test_easyswooleDocker(t *testing.T) {
	dockerfile := "./Dockerfile"
	err := easyswooleDocker(dockerfile)
	if err != nil {
		t.Fatal(err)
	} else {
		defer os.Remove(dockerfile)
	}

	var exist = true
	if _, err := os.Stat(dockerfile); os.IsNotExist(err) {
		exist = false
	}
	if !exist {
		t.Fatal("not match expected result")
	}
}
