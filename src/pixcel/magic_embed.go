// Copyright (c) 2026 H0llyW00dzZ All rights reserved.
//
// By accessing or using this software, you agree to be bound by the terms
// of the License Agreement, which you can find at LICENSE files.

package pixcel

import (
	_ "embed" // required for go:embed directive
	"html/template"
)

//go:embed template.go.tmpl
var pixelArtTemplate string

var tmpl = template.Must(template.New("pixelart").Parse(pixelArtTemplate))
