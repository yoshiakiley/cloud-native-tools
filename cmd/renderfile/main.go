package main

import (
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

var file string

func main() {
	flag.StringVar(&file, "f", "*.conf", "-f xx.conf")
	flag.Parse()

	rc, config, err := readAll(file)
	if err != nil {
		fmt.Println(err)
		return
	}
	if err := rc.Close(); err != nil {
		panic(err)
	}

	configTemplate := template.New("config")
	configTemplate = template.Must(configTemplate.Parse(string(config)))
	variables := findVariables(string(config))
	args := make(map[string]interface{})
	for _, key := range variables {
		value := os.Getenv(key)
		if value != "" {
			args[key] = value
		}
	}
	output := &output{}
	if err := configTemplate.Execute(output, args); err != nil {
		panic(err)
	}
	fmt.Printf("output--------------"+
		"%s\n", output.data)

	if err := ioutil.WriteFile(file, output.data, 0777); err != nil {
		panic(err)
	}
}

var _ io.Writer = &output{}

type output struct {
	data []byte
}

func (o *output) Write(p []byte) (int, error) {
	o.data = append(o.data, p...)
	return len(p), nil
}

func readAll(file string) (io.WriteCloser, []byte, error) {
	f, err := os.OpenFile(file, os.O_RDWR, 0777)
	if err != nil {
		return nil, []byte(""), err
	}
	b, err := ioutil.ReadAll(f)
	return f, b, err
}

func findVariables(data string) []string {
	pattern := `\{\{\.([A-Z])*\}\}`
	result := make([]string, 0)
	for _, key := range regexp.MustCompile(pattern).FindAllString(data, -1) {
		result = append(result, strings.TrimLeft(strings.TrimRight(strings.TrimLeft(key, "{"), "}"), "."))
	}
	return result
}
