package utils

import (
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

// ListDirectory list all path return the file name but not include file full path
// ignore the hide file
func ListDirectory(paths ...string) ([]string, error) {
	all := make([]string, 0)
	for _, path := range paths {
		files, err := ioutil.ReadDir(path)
		if err != nil {
			return nil, err
		}
		for _, file := range files {
			// ignore hide
			if strings.HasPrefix(file.Name(), ".") {
				continue
			}
			all = append(all, file.Name())
		}
	}
	return all, nil
}

func ReadAll(file string) (io.WriteCloser, []byte, error) {
	f, err := os.OpenFile(file, os.O_RDWR, 0777)
	if err != nil {
		return nil, []byte(""), err
	}
	b, err := ioutil.ReadAll(f)
	return f, b, err
}

func FindVariables(data string) []string {
	pattern := `\{\{\.([A-Z])*\}\}`
	result := make([]string, 0)
	for _, key := range regexp.MustCompile(pattern).FindAllString(data, -1) {
		result = append(result, strings.TrimLeft(strings.TrimRight(strings.TrimLeft(key, "{"), "}"), "."))
	}
	return result
}

func GenerateFile(filename string, content string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	_, err = io.WriteString(f, content)
	if err != nil {
		return err
	}
	return nil
}
