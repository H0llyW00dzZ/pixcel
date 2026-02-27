// Copyright (c) 2026 H0llyW00dzZ All rights reserved.
//
// By accessing or using this software, you agree to be bound by the terms
// of the License Agreement, which you can find at LICENSE files.

package pixcel

// Option is a functional option for configuring the Converter.
type Option func(*Converter)

// WithTargetWidth configures the output table width in cells (pixels).
func WithTargetWidth(w int) Option {
	return func(c *Converter) {
		if w > 0 {
			c.targetWidth = w
		}
	}
}

// WithHTMLWrapper determines if the output includes the full <html>, <head>, and <body>
// wrapper around the generated table. If false, only the <table> is output.
func WithHTMLWrapper(enabled bool, title string) Option {
	return func(c *Converter) {
		c.withHTML = enabled
		if title != "" {
			c.htmlTitle = title
		}
	}
}

// WithTargetHeight configures the output table height in cells (pixels).
// When set, it overrides the proportional height calculation.
// If both [WithTargetWidth] and WithTargetHeight are set, the image
// is stretched to exactly those dimensions.
func WithTargetHeight(h int) Option {
	return func(c *Converter) {
		if h > 0 {
			c.targetHeight = h
		}
	}
}
