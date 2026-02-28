# pixcel

[![Go Version](https://img.shields.io/badge/Go-%3E%3D1.25.7-blue?logo=go)](https://go.dev/dl/)
[![Go Reference](https://pkg.go.dev/badge/github.com/H0llyW00dzZ/pixcel.svg)](https://pkg.go.dev/github.com/H0llyW00dzZ/pixcel)
[![Go Report Card](https://goreportcard.com/badge/github.com/H0llyW00dzZ/pixcel)](https://goreportcard.com/report/github.com/H0llyW00dzZ/pixcel)
[![License: BSD-3-Clause](https://img.shields.io/badge/License-BSD--3--Clause-blue.svg)](LICENSE)
[![codecov](https://codecov.io/gh/H0llyW00dzZ/pixcel/branch/master/graph/badge.svg?token=EZEFQ7RDQP)](https://codecov.io/gh/H0llyW00dzZ/pixcel)

Convert images into HTML table-based pixel art.

<p align="center">
  <img src="./logo_by_grok_4.20-beta.jpg" alt="PIXCEL Logo" width="500">
  <br>
  <i>Image By Grok 4.20 (beta)</i>
</p>

**pixcel** is a Go SDK and CLI tool that transforms any PNG, JPEG, or GIF image into a self-contained HTML file using an optimised `<table>` layout with 2D greedy meshing (`colspan` + `rowspan`) for minimal DOM output.

## Performance

`pixcel` uses a **2D greedy meshing algorithm** to pack identical-color pixel regions into the fewest possible `<td>` elements by calculating both `colspan` (horizontal) and `rowspan` (vertical) spans.

Given an image matrix $I$ of dimensions $W \times H$, the algorithm finds maximal rectangles of uniform color:

$$w = \max \left\{ n \geq 1 \;\middle|\; \forall\, k \in [0, n): \; I(x_0+k,\, y_0) = c_0 \right\}$$

$$h = \max \left\{ m \geq 1 \;\middle|\; \forall\, j \in [0, m),\; \forall\, k \in [0, w): \; I(x_0+k,\, y_0+j) = c_0 \right\}$$

Each rectangle emits a single cell: `<td colspan="w" rowspan="h">`, achieving $O(W \times H)$ time complexity with up to **95%+ payload reduction** compared to naive pixel-per-cell output.

## Installation

### CLI

```bash
go install github.com/H0llyW00dzZ/pixcel/cmd/pixcel@latest
```

### SDK

> [!NOTE]
> The SDK requires **Go 1.25.7** or later.

```bash
go get github.com/H0llyW00dzZ/pixcel
```

## Usage

### CLI

```bash
# Basic conversion
pixcel convert photo.png

# Custom dimensions
pixcel convert logo.jpg -W 80 -o art.html

# Fixed width and height (stretches to exact dimensions)
pixcel convert icon.gif -W 100 -H 50 -o stretched.html

# Table-only mode (no HTML wrapper)
pixcel convert sprite.png --no-html -o table.html

# Custom page title
pixcel convert art.png -t "My Pixel Art" -o gallery.html

# Smooth load (hides content until fully rendered)
pixcel convert art.png -W 600 -H 306 -o art.html --smooth-load
```

### SDK

```go
package main

import (
    "context"
    "image"
    _ "image/png"
    "os"

    "github.com/H0llyW00dzZ/pixcel/src/pixcel"
)

func main() {
    // Open and decode the image
    f, _ := os.Open("photo.png")
    defer f.Close()
    img, _, _ := image.Decode(f)

    // Create a converter with options
    converter := pixcel.New(
        pixcel.WithTargetWidth(64),
        pixcel.WithTargetHeight(32),            // optional: fixed height
        pixcel.WithHTMLWrapper(true, "Art"),   // full HTML page with title
        pixcel.WithSmoothLoad(true),            // optional: hide until loaded
    )

    // Convert and write to file
    out, _ := os.Create("output.html")
    defer out.Close()
    converter.Convert(context.Background(), img, out)
}
```

## Options

| Option | CLI Flag | Default | Description |
|--------|----------|---------|-------------|
| `WithTargetWidth` | `-W, --width` | `56` | Output width in table cells |
| `WithTargetHeight` | `-H, --height` | proportional | Output height in table cells |
| `WithHTMLWrapper` | `--no-html` | `true` | Include full HTML document wrapper |
| `WithSmoothLoad` | `--smooth-load` | `false` | Hide content until fully loaded to prevent progressive rendering |
| — | `-t, --title` | `Go Pixel Art` | HTML page title |
| — | `-o, --output` | `go_pixel_art.html` | Output file path |

## Project Structure

```
pixcel/
├── cmd/pixcel/         # CLI entry point — minimal main
├── internal/cli/       # CLI layer — Cobra commands, flag binding
├── src/pixcel/         # Core SDK — Converter, options, HTML generation
├── .github/workflows/  # CI configuration
└── Makefile            # Test and build targets
```

## Testing

```bash
# Clone the repository
git clone https://github.com/H0llyW00dzZ/pixcel.git
cd pixcel

# Run tests with race detector
make test

# Run tests with verbose output
make test-verbose

# Run tests with coverage report
make test-cover

# View coverage in browser
go tool cover -html=coverage.txt

# Clean generated files
make clean
```

## License

[BSD-3-Clause](LICENSE) © H0llyW00dzZ
