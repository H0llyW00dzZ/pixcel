// Copyright (c) 2026 H0llyW00dzZ All rights reserved.
//
// By accessing or using this software, you agree to be bound by the terms
// of the License Agreement, which you can find at LICENSE files.

package pixcel

import (
	"context"
	"image"
	"io"
)

// Converter is the main client for converting an image to HTML pixel art.
// It uses functional options for configuring the conversion process.
type Converter struct {
	targetWidth  int
	targetHeight int
	withHTML     bool
	htmlTitle    string
}

// New creates a new Converter with the provided options.
// It applies default settings which can be overridden by the options.
func New(opts ...Option) *Converter {
	c := &Converter{
		targetWidth: 56,
		withHTML:    true,
		htmlTitle:   "Pixel Art",
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Convert takes an image and writes the HTML table pixel art to the provided writer.
//
// Convert returns [ErrNilImage] if img is nil, and [ErrNilWriter] if w is nil.
func (c *Converter) Convert(ctx context.Context, img image.Image, w io.Writer) error {
	if img == nil {
		return ErrNilImage
	}
	if w == nil {
		return ErrNilWriter
	}
	return c.generateHTML(ctx, img, w)
}
