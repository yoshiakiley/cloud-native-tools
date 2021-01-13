package main

import (
	"fmt"
	"io"
	"text/template"
)

var _ io.Writer = &Output{}

type Output struct{ Data []byte }

func (o *Output) Write(p []byte) (n int, err error) {
	o.Data = append(o.Data, p...)
	if len(o.Data) < 1 {
		err = fmt.Errorf("can't not copy")
	}
	return
}

func Render(data interface{}, tpl string) (string, error) {
	t, err := template.New("").Parse(tpl)
	if err != nil {
		return "", err
	}
	o := &Output{}
	if err := t.Execute(o, data); err != nil {
		return "", err
	}

	return string(o.Data), nil
}
