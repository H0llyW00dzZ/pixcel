// Copyright (c) 2026 H0llyW00dzZ All rights reserved.
//
// By accessing or using this software, you agree to be bound by the terms
// of the License Agreement, which you can find at LICENSE files.

package pixcel

import (
	"context"
	_ "embed" // required for go:embed directive
	"fmt"
	"image"
	"io"
	"math"

	"golang.org/x/image/draw"
)

// Cell represents a single `<td>` block with a specific color and column span.
type Cell struct {
	Color   string
	Colspan int
}

// templateData holds the dynamic data injected into the HTML template.
type templateData struct {
	WithHTML   bool
	Title      string
	Width      int
	Height     int
	Rows       [][]Cell
	SmoothLoad bool
}

// generateHTML contains the core logic for scaling the image and building
// the optimized HTML table output via templates.
func (c *Converter) generateHTML(ctx context.Context, img image.Image, w io.Writer) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	destImg, err := c.scaleImage(img)
	if err != nil {
		return err
	}

	data, err := c.buildTemplateData(ctx, destImg)
	if err != nil {
		return err
	}

	return tmpl.Execute(w, data)
}

// scaleImage scales the provided image to the converter's target dimensions.
// If targetHeight is set, it uses that value directly; otherwise it calculates
// height proportionally from targetWidth.
func (c *Converter) scaleImage(img image.Image) (image.Image, error) {
	bounds := img.Bounds()
	origW := bounds.Dx()
	origH := bounds.Dy()

	if origW == 0 || origH == 0 {
		return nil, ErrInvalidDimensions
	}

	targetW := c.targetWidth
	targetH := c.targetHeight
	if targetH == 0 {
		targetH = int(math.Round(float64(origH) * float64(targetW) / float64(origW)))
		if targetH == 0 {
			targetH = 1
		}
	}

	destImg := image.NewRGBA(image.Rect(0, 0, targetW, targetH))
	draw.NearestNeighbor.Scale(destImg, destImg.Bounds(), img, bounds, draw.Over, nil)
	return destImg, nil
}

// buildTemplateData constructs the data needed for the HTML template, computing colspans.
func (c *Converter) buildTemplateData(ctx context.Context, img image.Image) (*templateData, error) {
	bounds := img.Bounds()
	targetW := bounds.Max.X
	targetH := bounds.Max.Y

	rows := make([][]Cell, 0, targetH)

	for y := range targetH {
		// Periodically check context during table building block
		if y%10 == 0 {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
		}

		rows = append(rows, buildRow(img, y, targetW))
	}

	return &templateData{
		WithHTML:   c.withHTML,
		Title:      c.htmlTitle,
		Width:      targetW,
		Height:     targetH,
		Rows:       rows,
		SmoothLoad: c.smoothLoad,
	}, nil
}

// buildRow calculates consecutive color spans to optimize the table payload.
func buildRow(img image.Image, y, width int) []Cell {
	var row []Cell
	x := 0
	for x < width {
		r, g, b, _ := img.At(x, y).RGBA()
		r8, g8, b8 := uint8(r>>8), uint8(g>>8), uint8(b>>8)

		span := 1
		for x+span < width {
			nr, ng, nb, _ := img.At(x+span, y).RGBA()
			if uint8(nr>>8) == r8 && uint8(ng>>8) == g8 && uint8(nb>>8) == b8 {
				span++
			} else {
				break
			}
		}

		row = append(row, Cell{
			Color:   fmt.Sprintf("#%02x%02x%02x", r8, g8, b8),
			Colspan: span,
		})

		x += span
	}
	return row
}
