package main

import (
	"github.com/golang-module/carbon/v2"
	"github.com/sbani/go-humanizer/numbers"
	"strings"
	"text/template"
)

var templateString = `Welcome to a Testing server for Neo's Dummythiccness {{.NL}}
You are the {{.UserNum}} user today.{{.NL}}
Current local server time is: {{.Time}}`
var timeFmt = `Y-m-d H:i:s T` // https://github.com/golang-module/carbon#format-sign-table

type Substitutions struct {
	Time    string
	UserNum string
	NL      string
}

func templateCompileAndExec(input string) string {
	tmpl := template.Must(template.New("motd").Parse(input))

	return templateExec(tmpl)
}

func templateExec(input *template.Template) string {
	s := new(Substitutions)
	s.Time = carbon.Now().Format(timeFmt)
	s.UserNum = numbers.Ordinalize(69)
	s.NL = "\\n"

	out := new(strings.Builder)
	err := input.Execute(out, s)
	if err != nil {
		return ""
	}

	return out.String()
}
