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

	for y := 0; y < height; y++ {
		if y%10 == 0 {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
		}

		var currentRow []Cell

		for x := 0; x < width; x++ {
			if visited[y][x] {
				// The HTML table layout automatically pushes cells rightward if there
				// is a rowspan from a row above protruding into this space, so we
				// do not emit an empty placeholder cell here.
				continue
			}

			// Capture the color of the current anchor pixel
			r, g, b, _ := img.At(x, y).RGBA()
			r8, g8, b8 := uint8(r>>8), uint8(g>>8), uint8(b>>8)

			// 1. Expand Horizontally (Colspan)
			w := 1
			for x+w < width && !visited[y][x+w] {
				nr, ng, nb, _ := img.At(x+w, y).RGBA()
				if uint8(nr>>8) == r8 && uint8(ng>>8) == g8 && uint8(nb>>8) == b8 {
					w++
				} else {
					break
				}
			}

			// 2. Expand Vertically (Rowspan)
			// We can only increase height if *every* pixel in the new proposed row block
			// (from x to x+w-1) perfectly matches the anchor color AND is unvisited.
			h := 1
			canExpand := true
			for y+h < height && canExpand {
				for dx := 0; dx < w; dx++ {
					if visited[y+h][x+dx] {
						canExpand = false
						break
					}
					nr, ng, nb, _ := img.At(x+dx, y+h).RGBA()
					if uint8(nr>>8) != r8 || uint8(ng>>8) != g8 || uint8(nb>>8) != b8 {
						canExpand = false
						break
					}
				}
				if canExpand {
					h++
				}
			}

			// 3. Mark the discovered NxM block as visited
			for dy := 0; dy < h; dy++ {
				for dx := 0; dx < w; dx++ {
					visited[y+dy][x+dx] = true
				}
			}

			// 4. Emit the finalized Cell
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
