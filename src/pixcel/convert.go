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

// Cell represents a single `<td>` block with a color, column span, and row span.
type Cell struct {
	Color   string
	Colspan int
	Rowspan int
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

// buildTemplateData constructs the data needed for the HTML template, computing 2D colspan/rowspan packing.
func (c *Converter) buildTemplateData(ctx context.Context, img image.Image) (*templateData, error) {
	bounds := img.Bounds()
	targetW := bounds.Max.X
	targetH := bounds.Max.Y

	rows, err := buildTable(ctx, img, targetW, targetH)
	if err != nil {
		return nil, err
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

// buildTable applies a 2D greedy meshing algorithm to map the image into the fewest
// possible HTML table cells by dynamically calculating both colspan and rowspan.
func buildTable(ctx context.Context, img image.Image, width, height int) ([][]Cell, error) {
	visited := make([][]bool, height)
	for i := range visited {
		visited[i] = make([]bool, width)
	}

	rows := make([][]Cell, height)

	for y := range height {
		if y%10 == 0 {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
		}

		var currentRow []Cell

		for x := range width {
			if visited[y][x] {
				continue
			}

			r8, g8, b8 := colorAt(img, x, y)
			w := expandWidth(img, visited[y], x, y, width, r8, g8, b8)
			h := expandHeight(img, x, y, w, height, r8, g8, b8)
			markVisited(visited, x, y, w, h)

			currentRow = append(currentRow, Cell{
				Color:   fmt.Sprintf("#%02x%02x%02x", r8, g8, b8),
				Colspan: w,
				Rowspan: h,
			})
		}
		rows[y] = currentRow
	}

	return rows, nil
}

// colorAt returns the 8-bit RGB components of the pixel at (x, y).
func colorAt(img image.Image, x, y int) (uint8, uint8, uint8) {
	r, g, b, _ := img.At(x, y).RGBA()
	return uint8(r >> 8), uint8(g >> 8), uint8(b >> 8)
}

// expandWidth calculates the maximum horizontal span of consecutive pixels
// matching the anchor color (r8, g8, b8) starting at column x on row y.
func expandWidth(img image.Image, visitedRow []bool, x, y, width int, r8, g8, b8 uint8) int {
	w := 1
	for x+w < width && !visitedRow[x+w] {
		nr, ng, nb := colorAt(img, x+w, y)
		if nr != r8 || ng != g8 || nb != b8 {
			break
		}
		w++
	}
	return w
}

// expandHeight calculates how many rows below y share the exact same color strip
// of width w starting at column x, without overlapping already-visited cells.
func expandHeight(img image.Image, x, y, w, height int, r8, g8, b8 uint8) int {
	h := 1
	for y+h < height {
		if !rowMatchesColor(img, x, y+h, w, r8, g8, b8) {
			break
		}
		h++
	}
	return h
}

// rowMatchesColor checks whether every pixel in the range [x, x+w) on the given
// row at vertical position y matches the anchor color.
func rowMatchesColor(img image.Image, x, y, w int, r8, g8, b8 uint8) bool {
	for dx := range w {
		nr, ng, nb := colorAt(img, x+dx, y)
		if nr != r8 || ng != g8 || nb != b8 {
			return false
		}
	}
	return true
}

// markVisited flags all pixels in the rectangular block as visited.
func markVisited(visited [][]bool, x, y, w, h int) {
	for dy := range h {
		for dx := range w {
			visited[y+dy][x+dx] = true
		}
	}
}
