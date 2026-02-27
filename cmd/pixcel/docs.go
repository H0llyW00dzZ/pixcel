// Copyright (c) 2026 H0llyW00dzZ All rights reserved.
//
// By accessing or using this software, you agree to be bound by the terms
// of the License Agreement, which you can find at LICENSE files.

// pixcel converts images into HTML table-based pixel art.
//
// # Installation
//
// Install the latest version using go install:
//
//	go install github.com/H0llyW00dzZ/pixcel/cmd/pixcel@latest
//
// # Usage
//
// Convert an image to HTML pixel art:
//
//	pixcel convert photo.png
//	pixcel convert logo.jpg -W 80 -o art.html
//	pixcel convert icon.gif -W 100 -H 50 -o stretched.html
//	pixcel convert icon.gif --no-html
//	pixcel convert art.png -W 600 -H 306 -o art.html --smooth-load
//
// # Flags
//
//   - -W, --width       target width in table cells (default: 56)
//   - -H, --height      target height in table cells (default: proportional)
//   - -o, --output      output HTML file path (default: go_pixel_art.html)
//   - -t, --title       title for the HTML page (default: Go Pixel Art)
//   - --no-html         output only the <table>, omit the HTML wrapper
//   - --smooth-load     hide content until fully loaded to prevent progressive rendering
//
// # SDK Usage
//
// The underlying SDK can also be imported directly:
//
//	go get github.com/H0llyW00dzZ/pixcel
//
// Then use it in your code:
//
//	converter := pixcel.New(
//	    pixcel.WithTargetWidth(64),
//	    pixcel.WithTargetHeight(32),
//	)
//	err := converter.Convert(ctx, img, os.Stdout)
package main
