// Copyright (c) 2026 H0llyW00dzZ All rights reserved.
//
// By accessing or using this software, you agree to be bound by the terms
// of the License Agreement, which you can find at LICENSE files.

package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/H0llyW00dzZ/pixcel/src/pixcel"
	"github.com/spf13/cobra"
)

// convertCmd flag values, scoped to this file.
var (
	flagWidth  int
	flagHeight int
	flagOutput string
	flagNoHTML bool
	flagTitle  string
)

// convertCmd converts an image file to HTML pixel art.
var convertCmd = &cobra.Command{
	Use:   "convert <image>",
	Short: renderTemplate("convert.short"),
	Long:  renderTemplate("convert.long"),
	Args:  cobra.ExactArgs(1),
	RunE:  runConvert,
}

func init() {
	convertCmd.Flags().IntVarP(&flagWidth, "width", "W", 56, "target width in table cells")
	convertCmd.Flags().IntVarP(&flagHeight, "height", "H", 0, "target height in table cells (default: proportional)")
	convertCmd.Flags().StringVarP(&flagOutput, "output", "o", "pixel_art.html", "output HTML file path")
	convertCmd.Flags().BoolVar(&flagNoHTML, "no-html", false, "output only the <table>, omit the HTML wrapper")
	convertCmd.Flags().StringVarP(&flagTitle, "title", "t", "Go Pixel Art", "title for the HTML page")

	rootCmd.AddCommand(convertCmd)
}

// runConvert is the RunE handler for the convert subcommand.
func runConvert(_ *cobra.Command, args []string) error {
	imagePath := args[0]

	// Try animated GIF path first.
	if g, err := loadGIF(imagePath); err == nil && len(g.Image) > 1 {
		fmt.Printf("Loaded animated gif (%d frames) from %s\n", len(g.Image), imagePath)

		converter := pixcel.New(
			pixcel.WithTargetWidth(flagWidth),
			pixcel.WithTargetHeight(flagHeight),
			pixcel.WithHTMLWrapper(!flagNoHTML, flagTitle),
		)

		outFile, err := os.Create(flagOutput)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer outFile.Close()

		if err := converter.ConvertGIF(context.Background(), g, outFile); err != nil {
			return fmt.Errorf("conversion failed: %w", err)
		}

		fmt.Printf("Done! Saved animated HTML pixel art to %s\n", flagOutput)
		return nil
	}

	// Static image path (PNG, JPEG, single-frame GIF).
	img, format, err := loadImage(imagePath)
	if err != nil {
		return err
	}
	fmt.Printf("Loaded %s image from %s\n", format, imagePath)

	converter := pixcel.New(
		pixcel.WithTargetWidth(flagWidth),
		pixcel.WithTargetHeight(flagHeight),
		pixcel.WithHTMLWrapper(!flagNoHTML, flagTitle),
	)

	outFile, err := os.Create(flagOutput)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	if err := converter.Convert(context.Background(), img, outFile); err != nil {
		return fmt.Errorf("conversion failed: %w", err)
	}

	fmt.Printf("Done! Saved HTML pixel art to %s\n", flagOutput)
	return nil
}
