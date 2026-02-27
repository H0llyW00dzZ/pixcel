// Copyright (c) 2026 H0llyW00dzZ All rights reserved.
//
// By accessing or using this software, you agree to be bound by the terms
// of the License Agreement, which you can find at LICENSE files.

package cli

import (
	"bytes"
	_ "embed" // required for go:embed directive
	"text/template"
)

//go:embed template.go.tmpl
var cliTemplate string

// tmpl is the parsed CLI template containing all command descriptions.
var tmpl = template.Must(template.New("cli").Parse(cliTemplate))

// renderTemplate executes a named template block and returns the result as a string.
func renderTemplate(name string) string {
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, name, nil); err != nil {
		return ""
	}
	return buf.String()
}
