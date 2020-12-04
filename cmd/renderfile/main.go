package main

import (
	"flag"
	"fmt"
	"github.com/yametech/cloud-native-tools/pkg/utils"
	"html/template"
	"io"
	"io/ioutil"
	"os"
)

var file string

func main() {
	flag.StringVar(&file, "f", "*.conf", "-f nginx.conf")
	flag.Parse()
	if file == "" {
		fmt.Printf("render not working.\n")
		return
	}
	if err := render(file); err != nil {
		fmt.Printf("render file %s error %s\n", file, err)
		os.Exit(1)
	}
}

func render(file string) error {
	rc, config, err := utils.ReadAll(file)
	if err != nil {
		return err
	}
	if err := rc.Close(); err != nil {
		return err
	}
	configTemplate := template.New("config")
	configTemplate = template.Must(configTemplate.Parse(string(config)))
	variables := utils.FindVariables(string(config))
	args := make(map[string]interface{})
	for _, key := range variables {
		value := os.Getenv(key)
		if value != "" {
			args[key] = value
		}
	}
	output := &output{}
	if err := configTemplate.Execute(output, args); err != nil {
		return err
	}
	fmt.Printf("output--------------"+"%s\n", output.data)

	if err := ioutil.WriteFile(file, output.data, 0777); err != nil {
		return err
	}
	return nil
}

var _ io.Writer = &output{}

type output struct {
	data []byte
}

func (o *output) Write(p []byte) (int, error) {
	o.data = append(o.data, p...)
	return len(p), nil
}
