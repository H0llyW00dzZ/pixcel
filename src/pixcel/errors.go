// Copyright (c) 2026 H0llyW00dzZ All rights reserved.
//
// By accessing or using this software, you agree to be bound by the terms
// of the License Agreement, which you can find at LICENSE files.

package pixcel

import "errors"

var (
	// ErrNilImage is returned when a nil image is passed to Convert.
	ErrNilImage = errors.New("pixcel: image must not be nil")

	// ErrNilWriter is returned when a nil writer is passed to Convert.
	ErrNilWriter = errors.New("pixcel: writer must not be nil")

	// ErrInvalidDimensions is returned when an image has zero width or height.
	ErrInvalidDimensions = errors.New("pixcel: invalid image dimensions (zero width or height)")

	// ErrNilGIF is returned when a nil *gif.GIF is passed to ConvertGIF.
	ErrNilGIF = errors.New("pixcel: gif must not be nil")

	// ErrNoFrames is returned when a GIF contains zero frames.
	ErrNoFrames = errors.New("pixcel: gif contains no frames")
)
