// Copyright (c) 2026 H0llyW00dzZ All rights reserved.
//
// By accessing or using this software, you agree to be bound by the terms
// of the License Agreement, which you can find at LICENSE files.

// Package cli implements the pixcel command-line interface using [cobra].
//
// Command descriptions are maintained in template.go.tmpl and embedded
// at compile time via [embed]. This makes it easy to update help text
// without modifying Go source files.
//
// It is designed to be invoked from a minimal main function:
//
//	func main() { cli.Execute() }
//
// # Adding a New Command
//
// Create a new file (e.g. info.go) and register it in init():
//
//	var infoCmd = &cobra.Command{
//	    Use:   "info <image>",
//	    Short: renderTemplate("info.short"),
//	    Long:  renderTemplate("info.long"),
//	}
//	func init() { rootCmd.AddCommand(infoCmd) }
//
// Then add the corresponding named blocks to template.go.tmpl:
//
//	{{define "info.short"}}Show image metadata{{end}}
//	{{define "info.long"}}Display dimensions, format, and color info.{{end}}
package cli
