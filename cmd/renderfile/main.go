package main

import (
	"flag"
	"fmt"
	toolsutils "github.com/yametech/cloud-native-tools/pkg/utils"
	"html/template"
	"io"
	"io/ioutil"
	"os"
)

var file string
var dir string

func main() {
	flag.StringVar(&file, "f", "*.conf", "-f xx.conf")
	flag.Parse()

	rc, config, err := toolsutils.ReadAll(file)
	if err != nil {
		fmt.Println(err)
		return
	}
	if err := rc.Close(); err != nil {
		panic(err)
	}

	configTemplate := template.New("config")
	configTemplate = template.Must(configTemplate.Parse(string(config)))
	variables := toolsutils.FindVariables(string(config))
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
	fmt.Printf("output--------------"+"%s\n", output.data)

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
