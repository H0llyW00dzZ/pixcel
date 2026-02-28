// Copyright (c) 2026 H0llyW00dzZ All rights reserved.
//
// By accessing or using this software, you agree to be bound by the terms
// of the License Agreement, which you can find at LICENSE files.

package pixcel

import "golang.org/x/image/draw"

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

// WithSmoothLoad controls whether the generated HTML hides its content until
// the page is fully loaded, then reveals it with a smooth transition.
// This prevents the "drawing animation" caused by progressive rendering of
// large pixel-art tables. When enabled, a small inline script is added.
//
// This option only takes effect when [WithHTMLWrapper] is enabled.
func WithSmoothLoad(enabled bool) Option {
	return func(c *Converter) {
		c.smoothLoad = enabled
	}
}

// WithScaler configures the image scaling algorithm used during conversion.
// The default is [draw.NearestNeighbor], which preserves hard pixel edges.
// For smoother downscaling of photos and logos, use [draw.CatmullRom] or
// [draw.BiLinear].
func WithScaler(s draw.Scaler) Option {
	return func(c *Converter) {
		if s != nil {
			c.scaler = s
		}
	}
}

// WithObfuscation controls whether the generated HTML table uses randomized
// inline CSS styling formats (hex lower/upper/mixed-case, rgb(), hsl()) and
// randomized background-color property-name casing for each cell. This
// preserves the visual output exactly while making the underlying HTML source
// code highly resistant to automated scraping (The Nightmare Scenario for Bots or AI).
//
// Note: while a determined attacker can still bypass this by rendering the HTML
// in a headless browser and running OCR on the screenshot, the attack cost is
// significantly higher than against a plain image — a plain PNG CAPTCHA can be
// read by a vision AI in a single API call, whereas this approach requires a full
// browser runtime, a render cycle, and a screenshot before OCR can even begin.
func WithObfuscation(enabled bool) Option {
	return func(c *Converter) {
		c.obfuscate = enabled
	}
}

// WithMaxFrames sets the maximum number of frames to process when converting
// an animated GIF. If the GIF contains more frames than this limit, frames
// are sampled uniformly to stay within the budget while preserving the first
// and last frames. The default is 10.
//
// This prevents excessively large HTML output and keeps CSS @keyframes
// animation performant in browsers.
func WithMaxFrames(n int) Option {
	return func(c *Converter) {
		if n > 0 {
			c.maxFrames = n
		}
	}
}
