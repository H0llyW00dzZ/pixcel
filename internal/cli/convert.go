// Copyright (c) 2026 H0llyW00dzZ All rights reserved.
//
// By accessing or using this software, you agree to be bound by the terms
// of the License Agreement, which you can find at LICENSE files.

package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/H0llyW00dzZ/pixcel/src/pixcel"
	"github.com/spf13/cobra"
	"golang.org/x/image/draw"
)

// convertCmd flag values, scoped to this file.
var (
	flagWidth      int
	flagHeight     int
	flagOutput     string
	flagNoHTML     bool
	flagTitle      string
	flagSmoothLoad bool
	flagScaler     string
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
	convertCmd.Flags().StringVarP(&flagOutput, "output", "o", "go_pixel_art.html", "output HTML file path")
	convertCmd.Flags().BoolVar(&flagNoHTML, "no-html", false, "output only the <table>, omit the HTML wrapper")
	convertCmd.Flags().StringVarP(&flagTitle, "title", "t", "Go Pixel Art", "title for the HTML page")
	convertCmd.Flags().BoolVar(&flagSmoothLoad, "smooth-load", false, "hide content until fully loaded to prevent progressive rendering")
	convertCmd.Flags().StringVar(&flagScaler, "scaler", "nearest", "scaling algorithm: nearest, catmullrom, bilinear, approxbilinear")

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
			pixcel.WithSmoothLoad(flagSmoothLoad),
			pixcel.WithScaler(parseScaler(flagScaler)),
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
		pixcel.WithSmoothLoad(flagSmoothLoad),
		pixcel.WithScaler(parseScaler(flagScaler)),
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

// parseScaler maps a CLI flag string to a [draw.Scaler] implementation.
func parseScaler(name string) draw.Scaler {
	switch strings.ToLower(name) {
	case "catmullrom":
		return draw.CatmullRom
	case "bilinear":
		return draw.BiLinear
	case "approxbilinear":
		return draw.ApproxBiLinear
	default:
		return draw.NearestNeighbor
	}
}
