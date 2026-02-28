// Copyright (c) 2026 H0llyW00dzZ All rights reserved.
//
// By accessing or using this software, you agree to be bound by the terms
// of the License Agreement, which you can find at LICENSE files.

package pixcel

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/gif"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	xdraw "golang.org/x/image/draw"
)

func createTestImage() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	red := color.RGBA{R: 255, A: 255}
	blue := color.RGBA{B: 255, A: 255}

	for y := range 2 {
		for x := range 4 {
			img.Set(x, y, red)
		}
	}
	for y := 2; y < 4; y++ {
		for x := range 4 {
			img.Set(x, y, blue)
		}
	}
	return img
}

func createCheckerboardImage() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	black := color.RGBA{A: 255}
	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	for y := range 4 {
		for x := range 4 {
			if (x+y)%2 == 0 {
				img.Set(x, y, black)
			} else {
				img.Set(x, y, white)
			}
		}
	}
	return img
}

func TestConverter_Convert(t *testing.T) {
	img := createTestImage()
	converter := New(WithTargetWidth(4), WithHTMLWrapper(false, ""))
	var buf bytes.Buffer

	err := converter.Convert(context.Background(), img, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, `<table width="4" height="4"`)
	assert.Contains(t, output, `<td colspan="4" rowspan="2" style="width:4px;height:2px" bgcolor="#ff0000"></td>`)
	assert.Contains(t, output, `<td colspan="4" rowspan="2" style="width:4px;height:2px" bgcolor="#0000ff"></td>`)
}

func TestConverter_Convert_WithHTMLWrapper(t *testing.T) {
	img := createTestImage()
	converter := New(WithTargetWidth(4), WithHTMLWrapper(true, "Test Art"))
	var buf bytes.Buffer

	err := converter.Convert(context.Background(), img, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "<!DOCTYPE html>")
	assert.Contains(t, output, "<title>Test Art</title>")
	assert.Contains(t, output, "</html>")
	assert.Contains(t, output, "pixcel-container")
}

func TestConverter_Convert_NilImage(t *testing.T) {
	converter := New()
	var buf bytes.Buffer

	err := converter.Convert(context.Background(), nil, &buf)
	assert.ErrorIs(t, err, ErrNilImage)
}

func TestConverter_Convert_NilWriter(t *testing.T) {
	err := New().Convert(context.Background(), createTestImage(), nil)
	assert.ErrorIs(t, err, ErrNilWriter)
}

func TestConverter_Convert_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var buf bytes.Buffer
	err := New(WithTargetWidth(4)).Convert(ctx, createTestImage(), &buf)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestConverter_Convert_ZeroDimensionImage(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 0, 0))
	var buf bytes.Buffer

	err := New(WithTargetWidth(4)).Convert(context.Background(), img, &buf)
	assert.ErrorIs(t, err, ErrInvalidDimensions)
}

func TestConverter_Convert_CheckerboardNoColspan(t *testing.T) {
	converter := New(WithTargetWidth(4), WithHTMLWrapper(false, ""))
	var buf bytes.Buffer

	err := converter.Convert(context.Background(), createCheckerboardImage(), &buf)
	require.NoError(t, err)
	assert.NotContains(t, buf.String(), "colspan")
}

func TestConverter_Convert_LargeImageContextCheck(t *testing.T) {
	// Tall image triggers the y%10 context check branch
	img := image.NewRGBA(image.Rect(0, 0, 2, 25))
	green := color.RGBA{G: 255, A: 255}
	for y := range 25 {
		for x := range 2 {
			img.Set(x, y, green)
		}
	}

	var buf bytes.Buffer
	err := New(WithTargetWidth(2), WithHTMLWrapper(false, "")).Convert(context.Background(), img, &buf)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), `bgcolor="#00ff00"`)
}

func TestConverter_Convert_SinglePixelWide(t *testing.T) {
	// Image that scales to targetH=0, testing the targetH=1 floor
	img := image.NewRGBA(image.Rect(0, 0, 100, 1))
	for x := range 100 {
		img.Set(x, 0, color.RGBA{R: 128, G: 64, B: 32, A: 255})
	}

	var buf bytes.Buffer
	err := New(WithTargetWidth(10), WithHTMLWrapper(false, "")).Convert(context.Background(), img, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, `<table width="10"`)
	assert.Contains(t, output, `height="1"`)
}

func TestNew_Defaults(t *testing.T) {
	c := New()
	assert.Equal(t, 56, c.targetWidth)
	assert.True(t, c.withHTML)
	assert.Equal(t, "Go Pixel Art", c.htmlTitle)
}

func TestWithTargetWidth_Zero(t *testing.T) {
	c := New(WithTargetWidth(0))
	assert.Equal(t, 56, c.targetWidth, "zero width should be ignored")
}

func TestWithTargetWidth_Negative(t *testing.T) {
	c := New(WithTargetWidth(-10))
	assert.Equal(t, 56, c.targetWidth, "negative width should be ignored")
}

func TestWithHTMLWrapper_EmptyTitle(t *testing.T) {
	c := New(WithHTMLWrapper(true, ""))
	assert.Equal(t, "Go Pixel Art", c.htmlTitle, "empty title should keep default")
}

func TestWithHTMLWrapper_CustomTitle(t *testing.T) {
	c := New(WithHTMLWrapper(true, "Custom"))
	assert.Equal(t, "Custom", c.htmlTitle)
}

func TestWithHTMLWrapper_Disabled(t *testing.T) {
	c := New(WithHTMLWrapper(false, ""))
	assert.False(t, c.withHTML)
}

func TestGetVersion(t *testing.T) {
	v := GetVersion()
	assert.NotEmpty(t, v)
	assert.Equal(t, Version, v)
}

func TestConverter_Convert_TableOnlyNoWrapper(t *testing.T) {
	img := createTestImage()
	converter := New(WithTargetWidth(4), WithHTMLWrapper(false, ""))
	var buf bytes.Buffer

	require.NoError(t, converter.Convert(context.Background(), img, &buf))
	output := buf.String()

	// Should have inline style fallback, no container
	assert.Contains(t, output, `style="border-collapse:collapse;font-size:0;line-height:0"`)
	assert.NotContains(t, output, "<!DOCTYPE html>")
	assert.NotContains(t, output, "pixcel-container")
}

func TestBuildTable_SingleColor(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 5, 5))
	for y := range 5 {
		for x := range 5 {
			img.Set(x, y, color.RGBA{R: 10, G: 20, B: 30, A: 255})
		}
	}

	rows, err := buildTable(context.Background(), img, 5, 5)
	require.NoError(t, err)
	require.Len(t, rows, 5)

	// The first row should contain the single 5x5 cell
	require.Len(t, rows[0], 1)
	assert.Equal(t, 5, rows[0][0].Colspan)
	assert.Equal(t, 5, rows[0][0].Rowspan)
	assert.Equal(t, "#0a141e", rows[0][0].Color)

	// The other rows should be completely empty because they were consumed by rowspan
	for y := 1; y < 5; y++ {
		assert.Empty(t, rows[y])
	}
}

func TestBuildTable_AlternatingColors(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})
	img.Set(1, 0, color.RGBA{G: 255, A: 255})
	img.Set(0, 1, color.RGBA{B: 255, A: 255})
	img.Set(1, 1, color.RGBA{R: 10, G: 10, B: 10, A: 255})

	rows, err := buildTable(context.Background(), img, 2, 2)
	require.NoError(t, err)
	require.Len(t, rows, 2)

	require.Len(t, rows[0], 2)
	assert.Equal(t, 1, rows[0][0].Colspan)
	assert.Equal(t, 1, rows[0][0].Rowspan)
	assert.Equal(t, 1, rows[0][1].Colspan)
	assert.Equal(t, 1, rows[0][1].Rowspan)

	require.Len(t, rows[1], 2)
}

func TestScaleImage_Proportional(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 50))
	c := New(WithTargetWidth(10))

	scaled, err := c.scaleImage(img)
	require.NoError(t, err)

	bounds := scaled.Bounds()
	assert.Equal(t, 10, bounds.Dx())
	assert.Equal(t, 5, bounds.Dy())
}

func TestScaleImage_ZeroDimension(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 0, 10))
	c := New(WithTargetWidth(10))

	_, err := c.scaleImage(img)
	assert.ErrorIs(t, err, ErrInvalidDimensions)
}

func TestBuildTemplateData_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	c := New()

	_, err := c.buildTemplateData(ctx, img)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestScaleImage_VeryTallNarrow(t *testing.T) {
	// Width=1 height=1000 → tests targetH floor when scaled
	img := image.NewRGBA(image.Rect(0, 0, 1, 1000))
	c := New(WithTargetWidth(1))

	scaled, err := c.scaleImage(img)
	require.NoError(t, err)
	assert.Equal(t, 1, scaled.Bounds().Dx())
	assert.Equal(t, 1000, scaled.Bounds().Dy())
}

func TestGenerateHTML_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	img := createTestImage()
	c := New(WithTargetWidth(4))
	var buf bytes.Buffer

	err := c.generateHTML(ctx, img, &buf)
	assert.ErrorIs(t, err, context.Canceled)
}

type cancelOnBoundsImage struct {
	image.Image
	cancel context.CancelFunc
}

func (c *cancelOnBoundsImage) Bounds() image.Rectangle {
	if c.cancel != nil {
		c.cancel()
	}
	return c.Image.Bounds()
}

func TestGenerateHTML_CancellationDuringBuild(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	// cancel will be called when Bounds() is accessed during scaleImage

	img := &cancelOnBoundsImage{
		Image:  createTestImage(),
		cancel: cancel,
	}

	c := New(WithTargetWidth(4))
	var buf bytes.Buffer

	err := c.generateHTML(ctx, img, &buf)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestMultipleColspanPerRow(t *testing.T) {
	// Create image: 2 red, 2 blue, 2 green → 3 colspan groups
	img := image.NewRGBA(image.Rect(0, 0, 6, 1))
	red := color.RGBA{R: 255, A: 255}
	blue := color.RGBA{B: 255, A: 255}
	green := color.RGBA{G: 255, A: 255}
	img.Set(0, 0, red)
	img.Set(1, 0, red)
	img.Set(2, 0, blue)
	img.Set(3, 0, blue)
	img.Set(4, 0, green)
	img.Set(5, 0, green)

	converter := New(WithTargetWidth(6), WithHTMLWrapper(false, ""))
	var buf bytes.Buffer

	require.NoError(t, converter.Convert(context.Background(), img, &buf))
	output := buf.String()

	// It should now find exactly three 2x1 cells
	assert.Equal(t, 3, strings.Count(output, `colspan="2"`), "expected 3 colspan=2 groups")
}

func TestWithTargetHeight_Valid(t *testing.T) {
	c := New(WithTargetHeight(32))
	assert.Equal(t, 32, c.targetHeight)
}

func TestWithTargetHeight_Zero(t *testing.T) {
	c := New(WithTargetHeight(0))
	assert.Equal(t, 0, c.targetHeight, "zero height should be ignored")
}

func TestWithTargetHeight_Negative(t *testing.T) {
	c := New(WithTargetHeight(-5))
	assert.Equal(t, 0, c.targetHeight, "negative height should be ignored")
}

func TestScaleImage_FixedHeight(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 50))
	c := New(WithTargetWidth(10), WithTargetHeight(20))

	scaled, err := c.scaleImage(img)
	require.NoError(t, err)

	bounds := scaled.Bounds()
	assert.Equal(t, 10, bounds.Dx())
	assert.Equal(t, 20, bounds.Dy(), "height should be fixed at 20, not proportional")
}

func TestConverter_Convert_WithHeight(t *testing.T) {
	img := createTestImage()
	converter := New(
		WithTargetWidth(4),
		WithTargetHeight(8),
		WithHTMLWrapper(false, ""),
	)
	var buf bytes.Buffer

	require.NoError(t, converter.Convert(context.Background(), img, &buf))

	output := buf.String()
	assert.Contains(t, output, `height="8"`)
}

func TestConverter_Convert_DefaultHeight(t *testing.T) {
	c := New()
	assert.Equal(t, 0, c.targetHeight, "default targetHeight should be 0 (proportional)")
}

// errWriter always returns an error on Write to trigger template execute errors.
type errWriter struct{}

func (errWriter) Write([]byte) (int, error) {
	return 0, assert.AnError
}

func TestGenerateHTML_WriterError(t *testing.T) {
	img := createTestImage()
	c := New(WithTargetWidth(4), WithHTMLWrapper(false, ""))

	err := c.generateHTML(context.Background(), img, errWriter{})
	assert.Error(t, err)
}

func TestConverter_Convert_WriterError(t *testing.T) {
	img := createTestImage()
	converter := New(WithTargetWidth(4), WithHTMLWrapper(false, ""))

	err := converter.Convert(context.Background(), img, errWriter{})
	assert.Error(t, err)
}

// --- Animated GIF tests ---

// createTestGIF creates a multi-frame GIF with the given number of frames
// and per-frame delay in centiseconds. Each frame is 4x4 with a unique color.
func createTestGIF(frameCount, delay int) *gif.GIF {
	g := &gif.GIF{}
	g.Config.Width = 4
	g.Config.Height = 4

	for i := range frameCount {
		img := image.NewPaletted(image.Rect(0, 0, 4, 4), color.Palette{
			color.RGBA{R: uint8(i * 80 % 256), G: uint8(i * 60 % 256), B: uint8(i * 40 % 256), A: 255},
		})
		// Fill with palette index 0.
		for y := range 4 {
			for x := range 4 {
				img.SetColorIndex(x, y, 0)
			}
		}
		g.Image = append(g.Image, img)
		g.Delay = append(g.Delay, delay)
		g.Disposal = append(g.Disposal, gif.DisposalNone)
	}
	return g
}

func TestConvertGIF_MultiFrame(t *testing.T) {
	g := createTestGIF(3, 10)
	converter := New(WithTargetWidth(4), WithHTMLWrapper(true, "Animated Art"))
	var buf bytes.Buffer

	err := converter.ConvertGIF(context.Background(), g, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "<!DOCTYPE html>")
	assert.Contains(t, output, "<title>Animated Art</title>")
	assert.Contains(t, output, "@keyframes pixcel-anim")
	assert.Contains(t, output, "pixcel-stage")
	assert.Contains(t, output, "pixcel-frame")
	// Should have 3 frame divs.
	assert.Equal(t, 3, strings.Count(output, `class="pixcel-frame"`))
}

func TestConvertGIF_SingleFrame(t *testing.T) {
	g := createTestGIF(1, 10)
	converter := New(WithTargetWidth(4), WithHTMLWrapper(false, ""))
	var buf bytes.Buffer

	// Single-frame GIF should still work through ConvertGIF.
	err := converter.ConvertGIF(context.Background(), g, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Equal(t, 1, strings.Count(output, `class="pixcel-frame"`))
}

func TestConvertGIF_NilGIF(t *testing.T) {
	converter := New()
	var buf bytes.Buffer

	err := converter.ConvertGIF(context.Background(), nil, &buf)
	assert.ErrorIs(t, err, ErrNilGIF)
}

func TestConvertGIF_EmptyFrames(t *testing.T) {
	g := &gif.GIF{}
	converter := New()
	var buf bytes.Buffer

	err := converter.ConvertGIF(context.Background(), g, &buf)
	assert.ErrorIs(t, err, ErrNoFrames)
}

func TestConvertGIF_NilWriter(t *testing.T) {
	g := createTestGIF(2, 10)
	converter := New()

	err := converter.ConvertGIF(context.Background(), g, nil)
	assert.ErrorIs(t, err, ErrNilWriter)
}

func TestConvertGIF_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	g := createTestGIF(3, 10)
	converter := New(WithTargetWidth(4))
	var buf bytes.Buffer

	err := converter.ConvertGIF(ctx, g, &buf)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestConvertGIF_WriterError(t *testing.T) {
	g := createTestGIF(2, 10)
	converter := New(WithTargetWidth(4), WithHTMLWrapper(false, ""))

	err := converter.ConvertGIF(context.Background(), g, errWriter{})
	assert.Error(t, err)
}

func TestConvertGIF_WithHeight(t *testing.T) {
	g := createTestGIF(2, 10)
	converter := New(WithTargetWidth(4), WithTargetHeight(8), WithHTMLWrapper(false, ""))
	var buf bytes.Buffer

	require.NoError(t, converter.ConvertGIF(context.Background(), g, &buf))
	assert.Contains(t, buf.String(), `height="8"`)
}

func TestConvertGIF_KeyframesTiming(t *testing.T) {
	g := createTestGIF(2, 10) // 10 centiseconds = 0.1s per frame
	converter := New(WithTargetWidth(4), WithHTMLWrapper(true, "Timing"))
	var buf bytes.Buffer

	require.NoError(t, converter.ConvertGIF(context.Background(), g, &buf))

	output := buf.String()
	// Total duration = 0.2s
	assert.Contains(t, output, "0.200s")
	// First frame delay = 0s
	assert.Contains(t, output, `animation-delay: 0.000s`)
	// Second frame delay = 0.1s
	assert.Contains(t, output, `animation-delay: 0.100s`)
}

func TestConvertGIF_DisposalBackground(t *testing.T) {
	g := createTestGIF(2, 10)
	g.Disposal[0] = gif.DisposalBackground

	converter := New(WithTargetWidth(4), WithHTMLWrapper(false, ""))
	var buf bytes.Buffer

	err := converter.ConvertGIF(context.Background(), g, &buf)
	require.NoError(t, err)
	assert.Equal(t, 2, strings.Count(buf.String(), `class="pixcel-frame"`))
}

func TestConvertGIF_DisposalPrevious(t *testing.T) {
	g := createTestGIF(3, 10)
	g.Disposal[1] = gif.DisposalPrevious

	converter := New(WithTargetWidth(4), WithHTMLWrapper(false, ""))
	var buf bytes.Buffer

	err := converter.ConvertGIF(context.Background(), g, &buf)
	require.NoError(t, err)
	assert.Equal(t, 3, strings.Count(buf.String(), `class="pixcel-frame"`))
}

func TestGifDelay_ZeroDelay(t *testing.T) {
	// Delay of 0 should fall back to 0.1s (100ms) per browser convention.
	g := createTestGIF(2, 0)
	converter := New(WithTargetWidth(4), WithHTMLWrapper(true, "Zero Delay"))
	var buf bytes.Buffer

	require.NoError(t, converter.ConvertGIF(context.Background(), g, &buf))
	// Total duration = 2 * 0.1s = 0.2s
	assert.Contains(t, buf.String(), "0.200s")
}

func TestGifDelay_OutOfRange(t *testing.T) {
	// Test gifDelay when index exceeds Delay slice length.
	g := &gif.GIF{Delay: []int{10}}
	d := gifDelay(g, 5) // index out of range
	assert.InDelta(t, 0.1, d, 0.001)
}

func TestCompositeFrames_NoConfigDimensions(t *testing.T) {
	// GIF with zero Config dimensions — should fallback to first frame bounds.
	g := createTestGIF(2, 10)
	g.Config.Width = 0
	g.Config.Height = 0

	converter := New(WithTargetWidth(4))
	composited := converter.compositeFrames(g)
	assert.Len(t, composited, 2)
	assert.Equal(t, 4, composited[0].Bounds().Dx())
}

func TestConvertGIF_ZeroDimensionFrame(t *testing.T) {
	// GIF with zero-size frames should return ErrInvalidDimensions.
	g := &gif.GIF{}
	g.Config.Width = 0
	g.Config.Height = 0
	zeroFrame := image.NewPaletted(image.Rect(0, 0, 0, 0), color.Palette{color.Black})
	g.Image = append(g.Image, zeroFrame)
	g.Delay = append(g.Delay, 10)

	converter := New(WithTargetWidth(4))
	var buf bytes.Buffer
	err := converter.ConvertGIF(context.Background(), g, &buf)
	assert.ErrorIs(t, err, ErrInvalidDimensions)
}

type mockContext struct {
	context.Context
	cancelCount int
	cancelAt    int
}

func (m *mockContext) Err() error {
	m.cancelCount++
	if m.cancelCount > m.cancelAt {
		return context.Canceled
	}
	return nil
}

func TestConvertGIF_ManyFrames_ContextCancel(t *testing.T) {
	// Cancel before ConvertGIF is called
	g := createTestGIF(10, 10)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	converter := New(WithTargetWidth(4))
	var buf bytes.Buffer

	err := converter.ConvertGIF(ctx, g, &buf)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestConvertGIF_ContextCancel_InFrameLoop(t *testing.T) {
	g := createTestGIF(1, 10)
	ctx := &mockContext{
		Context:  context.Background(),
		cancelAt: 1, // fails at the second call (inside frame loop)
	}

	converter := New(WithTargetWidth(4))
	var buf bytes.Buffer

	err := converter.ConvertGIF(ctx, g, &buf)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestConvertGIF_ContextCancel_InBuildRows(t *testing.T) {
	g := createTestGIF(1, 10)
	ctx := &mockContext{
		Context:  context.Background(),
		cancelAt: 2, // fails at the third call (inside buildTable y=0)
	}

	converter := New(WithTargetWidth(4))
	var buf bytes.Buffer

	err := converter.ConvertGIF(ctx, g, &buf)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestConvertGIF_TargetHMightBeZero(t *testing.T) {
	// targetH would be 0 naturally, so it should be clamped to 1.
	g := createTestGIF(1, 10)
	g.Config.Width = 100
	g.Config.Height = 1
	// modify the frame to width 100
	frame := image.NewPaletted(image.Rect(0, 0, 100, 1), color.Palette{color.White})
	g.Image[0] = frame

	converter := New(WithTargetWidth(10), WithHTMLWrapper(false, ""))
	var buf bytes.Buffer

	err := converter.ConvertGIF(context.Background(), g, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, `height="1"`)
}

func TestBuildRows_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Image taller than 10 rows to trigger y%10 context check.
	img := image.NewRGBA(image.Rect(0, 0, 2, 20))
	converter := New()

	_, err := converter.buildRows(ctx, img)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestBuildKeyframes_ZeroFrames(t *testing.T) {
	g := &gif.GIF{}
	kf := buildKeyframes(0, g)
	assert.Nil(t, kf)
}

func TestConvertGIF_ProportionalHeight(t *testing.T) {
	// GIF that is 4x2, targetWidth=4, no targetHeight → should calculate proportionally.
	g := &gif.GIF{}
	g.Config.Width = 4
	g.Config.Height = 2
	frame := image.NewPaletted(image.Rect(0, 0, 4, 2), color.Palette{
		color.RGBA{R: 255, A: 255},
	})
	for y := range 2 {
		for x := range 4 {
			frame.SetColorIndex(x, y, 0)
		}
	}
	g.Image = append(g.Image, frame, frame)
	g.Delay = append(g.Delay, 10, 10)

	converter := New(WithTargetWidth(4), WithHTMLWrapper(false, ""))
	var buf bytes.Buffer

	require.NoError(t, converter.ConvertGIF(context.Background(), g, &buf))
	assert.Contains(t, buf.String(), `height="2"`)
}

// --- SmoothLoad tests ---

func TestWithSmoothLoad_Enabled(t *testing.T) {
	c := New(WithSmoothLoad(true))
	assert.True(t, c.smoothLoad)
}

func TestWithSmoothLoad_Disabled(t *testing.T) {
	c := New(WithSmoothLoad(false))
	assert.False(t, c.smoothLoad)
}

func TestNew_Defaults_SmoothLoad(t *testing.T) {
	c := New()
	assert.False(t, c.smoothLoad, "default smoothLoad should be false")
}

func TestConverter_Convert_WithSmoothLoad(t *testing.T) {
	img := createTestImage()
	converter := New(WithTargetWidth(4), WithHTMLWrapper(true, "Smooth"), WithSmoothLoad(true))
	var buf bytes.Buffer

	require.NoError(t, converter.Convert(context.Background(), img, &buf))

	output := buf.String()
	assert.Contains(t, output, "opacity: 0")
	assert.Contains(t, output, "transition: opacity")
	assert.Contains(t, output, `.loaded`)
	assert.Contains(t, output, `<script>`)
}

func TestConverter_Convert_WithoutSmoothLoad(t *testing.T) {
	img := createTestImage()
	converter := New(WithTargetWidth(4), WithHTMLWrapper(true, "No Smooth"), WithSmoothLoad(false))
	var buf bytes.Buffer

	require.NoError(t, converter.Convert(context.Background(), img, &buf))

	output := buf.String()
	assert.NotContains(t, output, "opacity: 0")
	assert.NotContains(t, output, "transition: opacity")
	assert.NotContains(t, output, `<script>`)
}

func TestConvertGIF_WithSmoothLoad(t *testing.T) {
	g := createTestGIF(2, 10)
	converter := New(WithTargetWidth(4), WithHTMLWrapper(true, "Smooth GIF"), WithSmoothLoad(true))
	var buf bytes.Buffer

	require.NoError(t, converter.ConvertGIF(context.Background(), g, &buf))

	output := buf.String()
	assert.Contains(t, output, "visibility: hidden")
	assert.Contains(t, output, `.loaded`)
	assert.Contains(t, output, `<script>`)
}

func TestConvertGIF_WithoutSmoothLoad(t *testing.T) {
	g := createTestGIF(2, 10)
	converter := New(WithTargetWidth(4), WithHTMLWrapper(true, "No Smooth GIF"), WithSmoothLoad(false))
	var buf bytes.Buffer

	require.NoError(t, converter.ConvertGIF(context.Background(), g, &buf))

	output := buf.String()
	assert.NotContains(t, output, "visibility: hidden")
	assert.NotContains(t, output, `<script>`)
}

// --- Scaler tests ---

func TestWithScaler_Default(t *testing.T) {
	c := New()
	assert.NotNil(t, c.scaler, "default scaler should not be nil")
}

func TestWithScaler_CatmullRom(t *testing.T) {
	c := New(WithScaler(xdraw.CatmullRom))
	assert.NotNil(t, c.scaler)
}

func TestWithScaler_Nil(t *testing.T) {
	c := New(WithScaler(nil))
	assert.NotNil(t, c.scaler, "nil scaler should be ignored, keeping default")
}

func TestConverter_Convert_WithScaler(t *testing.T) {
	img := createTestImage()
	converter := New(
		WithTargetWidth(4),
		WithHTMLWrapper(false, ""),
		WithScaler(xdraw.CatmullRom),
	)
	var buf bytes.Buffer

	err := converter.Convert(context.Background(), img, &buf)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), `<table`)
}

func TestConvertGIF_WithScaler(t *testing.T) {
	g := createTestGIF(2, 10)
	converter := New(
		WithTargetWidth(4),
		WithHTMLWrapper(false, ""),
		WithScaler(xdraw.BiLinear),
	)
	var buf bytes.Buffer

	err := converter.ConvertGIF(context.Background(), g, &buf)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), `class="pixcel-frame"`)
}
