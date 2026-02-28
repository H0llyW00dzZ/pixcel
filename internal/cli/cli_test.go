// Copyright (c) 2026 H0llyW00dzZ All rights reserved.
//
// By accessing or using this software, you agree to be bound by the terms
// of the License Agreement, which you can find at LICENSE files.

package cli

import (
	"image"
	"image/color"
	"image/gif"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestPNG writes a minimal 4x4 red PNG to the given path.
func createTestPNG(t *testing.T, path string) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	red := color.RGBA{R: 255, A: 255}
	for y := range 4 {
		for x := range 4 {
			img.Set(x, y, red)
		}
	}

	f, err := os.Create(path)
	require.NoError(t, err)
	defer f.Close()

	require.NoError(t, png.Encode(f, img))
}

// createTestGIFFile writes a minimal animated GIF with the given number of frames.
func createTestGIFFile(t *testing.T, path string, frameCount int) {
	t.Helper()
	g := &gif.GIF{}
	g.Config.Width = 4
	g.Config.Height = 4

	for i := range frameCount {
		img := image.NewPaletted(image.Rect(0, 0, 4, 4), color.Palette{
			color.RGBA{R: uint8(i * 80 % 256), G: uint8(i * 60 % 256), A: 255},
		})
		for y := range 4 {
			for x := range 4 {
				img.SetColorIndex(x, y, 0)
			}
		}
		g.Image = append(g.Image, img)
		g.Delay = append(g.Delay, 10)
		g.Disposal = append(g.Disposal, gif.DisposalNone)
	}

	f, err := os.Create(path)
	require.NoError(t, err)
	defer f.Close()

	require.NoError(t, gif.EncodeAll(f, g))
}

func TestLoadImage_Success(t *testing.T) {
	dir := t.TempDir()
	imgPath := filepath.Join(dir, "test.png")
	createTestPNG(t, imgPath)

	img, format, err := loadImage(imgPath)
	require.NoError(t, err)
	assert.NotNil(t, img)
	assert.Equal(t, "png", format)
}

func TestLoadImage_FileNotFound(t *testing.T) {
	_, _, err := loadImage("/nonexistent/path/image.png")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open image")
}

func TestLoadImage_InvalidImage(t *testing.T) {
	dir := t.TempDir()
	badFile := filepath.Join(dir, "bad.png")
	require.NoError(t, os.WriteFile(badFile, []byte("not an image"), 0644))

	_, _, err := loadImage(badFile)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode image")
}

func TestRunConvert_Success(t *testing.T) {
	dir := t.TempDir()
	imgPath := filepath.Join(dir, "test.png")
	createTestPNG(t, imgPath)
	outPath := filepath.Join(dir, "output.html")

	flagWidth = 4
	flagOutput = outPath
	flagNoHTML = false
	flagTitle = "Test"

	err := runConvert(nil, []string{imgPath})
	require.NoError(t, err)

	info, err := os.Stat(outPath)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(0))
}

func TestRunConvert_BadImagePath(t *testing.T) {
	flagWidth = 4
	flagOutput = filepath.Join(t.TempDir(), "out.html")
	flagNoHTML = false
	flagTitle = "Test"

	err := runConvert(nil, []string{"/nonexistent/image.png"})
	assert.Error(t, err)
}

func TestRunConvert_BadOutputPath(t *testing.T) {
	dir := t.TempDir()
	imgPath := filepath.Join(dir, "test.png")
	createTestPNG(t, imgPath)

	flagWidth = 4
	flagOutput = "/nonexistent/dir/output.html"
	flagNoHTML = false
	flagTitle = "Test"

	err := runConvert(nil, []string{imgPath})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create output file")
}

func TestRunConvert_NoHTMLMode(t *testing.T) {
	dir := t.TempDir()
	imgPath := filepath.Join(dir, "test.png")
	createTestPNG(t, imgPath)
	outPath := filepath.Join(dir, "table_only.html")

	flagWidth = 4
	flagOutput = outPath
	flagNoHTML = true
	flagTitle = "Ignored"

	require.NoError(t, runConvert(nil, []string{imgPath}))

	data, err := os.ReadFile(outPath)
	require.NoError(t, err)
	assert.NotContains(t, string(data), "<!DOCTYPE html>")
}

func TestRunConvert_HTMLMode(t *testing.T) {
	dir := t.TempDir()
	imgPath := filepath.Join(dir, "test.png")
	createTestPNG(t, imgPath)
	outPath := filepath.Join(dir, "full.html")

	flagWidth = 4
	flagOutput = outPath
	flagNoHTML = false
	flagTitle = "My Art"

	require.NoError(t, runConvert(nil, []string{imgPath}))

	data, err := os.ReadFile(outPath)
	require.NoError(t, err)

	content := string(data)
	assert.Contains(t, content, "<!DOCTYPE html>")
	assert.Contains(t, content, "<title>My Art</title>")
}

func TestExecute_ConvertSubcommand(t *testing.T) {
	dir := t.TempDir()
	imgPath := filepath.Join(dir, "test.png")
	createTestPNG(t, imgPath)
	outPath := filepath.Join(dir, "exec_output.html")

	rootCmd.SetArgs([]string{"convert", imgPath, "-W", "4", "-o", outPath})
	Execute()

	data, err := os.ReadFile(outPath)
	require.NoError(t, err)
	assert.True(t, len(data) > 0)
}

func TestExecute_ConvertSubcommand_NoArgs(t *testing.T) {
	rootCmd.SetArgs([]string{"convert"})
	// Cobra will print an error but not exit when using SetArgs in tests
	err := rootCmd.Execute()
	assert.Error(t, err)
}

func TestRunConvert_VerifyConvertedContent(t *testing.T) {
	dir := t.TempDir()
	imgPath := filepath.Join(dir, "test.png")
	createTestPNG(t, imgPath)
	outPath := filepath.Join(dir, "verify.html")

	flagWidth = 4
	flagOutput = outPath
	flagNoHTML = false
	flagTitle = "Verify"

	require.NoError(t, runConvert(nil, []string{imgPath}))

	data, err := os.ReadFile(outPath)
	require.NoError(t, err)

	content := string(data)
	// Red image → all cells should be #ff0000
	assert.Contains(t, content, `bgcolor="#ff0000"`)
	// 4-wide uniform color → colspan=4
	assert.True(t, strings.Contains(content, `colspan="4"`))
}

func TestRunConvert_WithHeight(t *testing.T) {
	dir := t.TempDir()
	imgPath := filepath.Join(dir, "test.png")
	createTestPNG(t, imgPath)
	outPath := filepath.Join(dir, "height_output.html")

	flagWidth = 4
	flagHeight = 8
	flagOutput = outPath
	flagNoHTML = false
	flagTitle = "Height Test"

	require.NoError(t, runConvert(nil, []string{imgPath}))

	data, err := os.ReadFile(outPath)
	require.NoError(t, err)

	content := string(data)
	assert.Contains(t, content, `height="8"`)

	// Reset for other tests
	flagHeight = 0
}

func TestExecute_Version(t *testing.T) {
	rootCmd.SetArgs([]string{"--version"})
	Execute()
	assert.NotEmpty(t, rootCmd.Version)
}

func TestExecute_ConvertWithHeight(t *testing.T) {
	dir := t.TempDir()
	imgPath := filepath.Join(dir, "test.png")
	createTestPNG(t, imgPath)
	outPath := filepath.Join(dir, "exec_height.html")

	rootCmd.SetArgs([]string{"convert", imgPath, "-W", "4", "--height", "10", "-o", outPath})
	Execute()

	data, err := os.ReadFile(outPath)
	require.NoError(t, err)
	assert.Contains(t, string(data), `height="10"`)
}

// --- Animated GIF CLI tests ---

func TestLoadGIF_Success(t *testing.T) {
	dir := t.TempDir()
	gifPath := filepath.Join(dir, "animated.gif")
	createTestGIFFile(t, gifPath, 3)

	g, err := loadGIF(gifPath)
	require.NoError(t, err)
	assert.Len(t, g.Image, 3)
	assert.Len(t, g.Delay, 3)
}

func TestLoadGIF_FileNotFound(t *testing.T) {
	_, err := loadGIF("/nonexistent/path/animated.gif")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open gif")
}

func TestLoadGIF_InvalidFile(t *testing.T) {
	dir := t.TempDir()
	badFile := filepath.Join(dir, "bad.gif")
	require.NoError(t, os.WriteFile(badFile, []byte("not a gif"), 0644))

	_, err := loadGIF(badFile)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode gif")
}

func TestRunConvert_AnimatedGIF(t *testing.T) {
	dir := t.TempDir()
	gifPath := filepath.Join(dir, "animated.gif")
	createTestGIFFile(t, gifPath, 3)
	outPath := filepath.Join(dir, "animated_output.html")

	flagWidth = 4
	flagHeight = 0
	flagOutput = outPath
	flagNoHTML = false
	flagTitle = "Animated Test"

	require.NoError(t, runConvert(nil, []string{gifPath}))

	data, err := os.ReadFile(outPath)
	require.NoError(t, err)

	content := string(data)
	assert.Contains(t, content, "@keyframes pixcel-anim")
	assert.Contains(t, content, "pixcel-frame")
	assert.Equal(t, 3, strings.Count(content, `class="pixcel-frame"`))
}

func TestRunConvert_SingleFrameGIF(t *testing.T) {
	dir := t.TempDir()
	gifPath := filepath.Join(dir, "single.gif")
	createTestGIFFile(t, gifPath, 1)
	outPath := filepath.Join(dir, "single_output.html")

	flagWidth = 4
	flagHeight = 0
	flagOutput = outPath
	flagNoHTML = false
	flagTitle = "Single Frame"

	// Single-frame GIF should fall through to static path.
	require.NoError(t, runConvert(nil, []string{gifPath}))

	data, err := os.ReadFile(outPath)
	require.NoError(t, err)

	content := string(data)
	// Static path — no animation CSS.
	assert.NotContains(t, content, "@keyframes")
}

func TestExecute_ConvertAnimatedGIF(t *testing.T) {
	dir := t.TempDir()
	gifPath := filepath.Join(dir, "animated.gif")
	createTestGIFFile(t, gifPath, 2)
	outPath := filepath.Join(dir, "exec_animated.html")

	rootCmd.SetArgs([]string{"convert", gifPath, "-W", "4", "-o", outPath})
	Execute()

	data, err := os.ReadFile(outPath)
	require.NoError(t, err)
	assert.Contains(t, string(data), "@keyframes pixcel-anim")
}

// --- Scaler CLI tests ---

func TestParseScaler(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"nearest", "nearest"},
		{"catmullrom", "catmullrom"},
		{"bilinear", "bilinear"},
		{"approxbilinear", "approxbilinear"},
		{"unknown", "nearest"}, // default fallback
	}
	for _, tt := range tests {
		s := parseScaler(tt.input)
		assert.NotNil(t, s, "parseScaler(%q) should not return nil", tt.input)
	}
}

func TestRunConvert_WithScaler(t *testing.T) {
	dir := t.TempDir()
	imgPath := filepath.Join(dir, "test.png")
	createTestPNG(t, imgPath)
	outPath := filepath.Join(dir, "scaler_output.html")

	flagWidth = 4
	flagHeight = 0
	flagOutput = outPath
	flagNoHTML = false
	flagTitle = "Scaler Test"
	flagSmoothLoad = false
	flagScaler = "catmullrom"

	require.NoError(t, runConvert(nil, []string{imgPath}))

	data, err := os.ReadFile(outPath)
	require.NoError(t, err)
	assert.Contains(t, string(data), `<table`)

	// Reset
	flagScaler = "nearest"
}
