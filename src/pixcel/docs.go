// Copyright (c) 2026 H0llyW00dzZ All rights reserved.
//
// By accessing or using this software, you agree to be bound by the terms
// of the License Agreement, which you can find at LICENSE files.

// Package pixcel converts images into HTML table-based pixel art.
//
// It uses the [Converter] type, configured via functional options, to scale an
// input [image.Image] and render it as an optimized HTML <table> with colspan
// merging for consecutive same-color cells.
//
// # Quick Start
//
//	converter := pixcel.New(
//	    pixcel.WithTargetWidth(64),
//	)
//	err := converter.Convert(ctx, img, os.Stdout)
//
// # Options
//
//   - [WithTargetWidth] sets the output width in table cells (default: 56).
//   - [WithTargetHeight] sets the output height in table cells (default: proportional).
//   - [WithHTMLWrapper] toggles the full HTML document wrapper (default: on).
//   - [WithSmoothLoad] hides content until fully loaded to prevent progressive rendering (default: off).
//   - [WithScaler] sets the image scaling algorithm: NearestNeighbor, CatmullRom, BiLinear, ApproxBiLinear (default: NearestNeighbor).
//   - [WithObfuscation] randomises inline CSS color formats (hex/rgb/hsl) and property-name casing for bot resistance (default: off; browser use only).
package pixcel
