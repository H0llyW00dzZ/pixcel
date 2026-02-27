// Copyright (c) 2026 H0llyW00dzZ All rights reserved.
//
// By accessing or using this software, you agree to be bound by the terms
// of the License Agreement, which you can find at LICENSE files.

package cli

import (
	"os"

	"github.com/H0llyW00dzZ/pixcel/src/pixcel"
	"github.com/spf13/cobra"
)

// rootCmd is the top-level Cobra command for pixcel.
var rootCmd = &cobra.Command{
	Use:   "pixcel",
	Short: renderTemplate("root.short"),
	Long:  renderTemplate("root.long"),
}

// Execute runs the root command.
func Execute() {
	rootCmd.Version = pixcel.GetVersion()

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
