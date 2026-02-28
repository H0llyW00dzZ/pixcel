// Copyright (c) 2026 H0llyW00dzZ All rights reserved.
//
// By accessing or using this software, you agree to be bound by the terms
// of the License Agreement, which you can find at LICENSE files.

package pixcel

import (
	"context"
	_ "embed" // required for go:embed directive
	"fmt"
	"html"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"io"
	"math"

	xdraw "golang.org/x/image/draw"
)

// gifFrameData holds the rendered rows and animation keyframes for a single GIF frame.
type gifFrameData struct {
	Rows      [][]Cell
	Keyframes []gifKeyframe
}

// gifKeyframe represents a single step in the CSS @keyframes rule.
type gifKeyframe struct {
	Percent string
	Opacity int
}

// gifTemplateData holds all data injected into the animated GIF HTML template.
type gifTemplateData struct {
	WithHTML         bool
	Title            string
	Width            int
	Height           int
	TotalDurationCSS string
	Frames           []gifFrameData
	SmoothLoad       bool
	Obfuscate        bool
}

// ConvertGIF takes an animated GIF and writes animated HTML pixel art to the
// provided writer. Each frame becomes a separate table layer, animated with
// pure CSS @keyframes.
//
// ConvertGIF returns [ErrNilGIF] if g is nil, [ErrNilWriter] if w is nil,
// and [ErrNoFrames] if the GIF contains no frames.
func (c *Converter) ConvertGIF(ctx context.Context, g *gif.GIF, w io.Writer) error {
	if g == nil {
		return ErrNilGIF
	}
	if w == nil {
		return ErrNilWriter
	}
	if len(g.Image) == 0 {
		return ErrNoFrames
	}
	return c.generateGIFHTML(ctx, g, w)
}

// generateGIFHTML composites GIF frames, scales them, and renders the animated
// HTML output via template.
func (c *Converter) generateGIFHTML(ctx context.Context, g *gif.GIF, w io.Writer) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// Composite all frames into full images (handling GIF disposal).
	composited := c.compositeFrames(g)

	// Calculate target dimensions from the first frame.
	firstBounds := composited[0].Bounds()
	origW := firstBounds.Dx()
	origH := firstBounds.Dy()

	if origW == 0 || origH == 0 {
		return ErrInvalidDimensions
	}

	targetW := c.targetWidth
	targetH := c.targetHeight
	if targetH == 0 {
		targetH = int(math.Round(float64(origH) * float64(targetW) / float64(origW)))
		if targetH == 0 {
			targetH = 1
		}
	}

	// Build CSS keyframes for all frames.
	allKeyframes := buildAllKeyframes(len(composited), g)

	// Build frame data.
	frames := make([]gifFrameData, 0, len(composited))
	var cumulativeDelay float64

	for i, img := range composited {
		if i%5 == 0 {
			if err := ctx.Err(); err != nil {
				return err
			}
		}

		scaled := c.scaleToSize(img, targetW, targetH)

		rows, err := c.buildRows(ctx, scaled)
		if err != nil {
			return err
		}

		delay := gifDelay(g, i)

		frames = append(frames, gifFrameData{
			Rows:      rows,
			Keyframes: allKeyframes[i],
		})

		cumulativeDelay += delay
	}

	data := &gifTemplateData{
		WithHTML:         c.withHTML,
		Title:            html.EscapeString(c.htmlTitle),
		Width:            targetW,
		Height:           targetH,
		TotalDurationCSS: fmt.Sprintf("%.3fs", cumulativeDelay),
		Frames:           frames,
		SmoothLoad:       c.smoothLoad,
	}

	return gifTmpl.Execute(w, data)
}

// compositeFrames renders each GIF frame onto a full-size canvas, handling
// the GIF disposal method to produce complete images for each frame.
func (c *Converter) compositeFrames(g *gif.GIF) []*image.RGBA {
	// Use the overall GIF dimensions as the canvas size.
	canvasW := g.Config.Width
	canvasH := g.Config.Height
	if canvasW == 0 || canvasH == 0 {
		// Fallback: use first frame bounds.
		b := g.Image[0].Bounds()
		canvasW = b.Max.X
		canvasH = b.Max.Y
	}

	canvas := image.NewRGBA(image.Rect(0, 0, canvasW, canvasH))
	result := make([]*image.RGBA, 0, len(g.Image))

	for i, frame := range g.Image {
		// Draw this frame onto the canvas.
		draw.Draw(canvas, frame.Bounds(), frame, frame.Bounds().Min, draw.Over)

		// Snapshot the current canvas.
		snapshot := image.NewRGBA(canvas.Bounds())
		copy(snapshot.Pix, canvas.Pix)
		result = append(result, snapshot)

		// Handle disposal.
		if i < len(g.Disposal) {
			switch g.Disposal[i] {
			case gif.DisposalBackground:
				// Clear the frame area to transparent.
				draw.Draw(canvas, frame.Bounds(),
					image.NewUniform(color.Transparent), image.Point{}, draw.Src)
			case gif.DisposalPrevious:
				// Restore to previous frame (re-copy previous snapshot).
				if i > 0 {
					copy(canvas.Pix, result[i-1].Pix)
				}
			}
			// DisposalNone (0) or default: leave canvas as-is.
		}
	}

	return result
}

// scaleToSize scales an image to the given target dimensions.
func (c *Converter) scaleToSize(img image.Image, targetW, targetH int) *image.RGBA {
	destImg := image.NewRGBA(image.Rect(0, 0, targetW, targetH))
	c.scaler.Scale(destImg, destImg.Bounds(), img, img.Bounds(), xdraw.Over, nil)
	return destImg
}

// buildRows creates the cell rows from a scaled image via 2D meshing.
func (c *Converter) buildRows(ctx context.Context, img image.Image) ([][]Cell, error) {
	bounds := img.Bounds()
	h := bounds.Max.Y
	w := bounds.Max.X

	return buildTable(ctx, img, w, h, c.obfuscate)
}

// gifDelay returns the delay for frame i in seconds.
// GIF delays are in centiseconds (1/100s). A delay of 0 is treated as 100ms
// per common browser behavior.
func gifDelay(g *gif.GIF, i int) float64 {
	if i < len(g.Delay) && g.Delay[i] > 0 {
		return float64(g.Delay[i]) / 100.0
	}
	return 0.1 // default 100ms
}

// buildAllKeyframes generates absolute-timed CSS @keyframes for each frame.
func buildAllKeyframes(frameCount int, g *gif.GIF) [][]gifKeyframe {
	if frameCount == 0 {
		return nil
	}

	// Calculate total duration.
	var totalDelay float64
	for i := range frameCount {
		totalDelay += gifDelay(g, i)
	}

	var result [][]gifKeyframe
	var cumulative float64

	for i := range frameCount {
		onPct := cumulative / totalDelay * 100
		delay := gifDelay(g, i)
		offPct := (cumulative + delay) / totalDelay * 100

		var keyframes []gifKeyframe

		// Start hidden if not the very first frame to appear.
		if onPct > 0 {
			keyframes = append(keyframes, gifKeyframe{Percent: "0%", Opacity: 0})
		}

		// Show frame.
		keyframes = append(keyframes, gifKeyframe{Percent: fmt.Sprintf("%.4f%%", onPct), Opacity: 1})

		// Hide frame when its delay expires.
		if offPct < 100 {
			keyframes = append(keyframes, gifKeyframe{Percent: fmt.Sprintf("%.4f%%", offPct), Opacity: 0})
		}

		// End: last frame stays visible until the loop restarts; others hide.
		if i == frameCount-1 {
			keyframes = append(keyframes, gifKeyframe{Percent: "100%", Opacity: 1})
		} else {
			keyframes = append(keyframes, gifKeyframe{Percent: "100%", Opacity: 0})
		}

		result = append(result, keyframes)
		cumulative += delay
	}

	return result
}
