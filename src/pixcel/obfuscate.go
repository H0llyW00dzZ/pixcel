// Copyright (c) 2026 H0llyW00dzZ All rights reserved.
//
// By accessing or using this software, you agree to be bound by the terms
// of the License Agreement, which you can find at LICENSE files.

package pixcel

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math"
	"strings"
)

// randIntn returns a cryptographically secure random integer in [0, n).
func randIntn(n int) int {
	if n <= 0 {
		return 0
	}
	var b [8]byte
	_, _ = rand.Read(b[:])
	return int(binary.LittleEndian.Uint64(b[:]) % uint64(n))
}

// formatColor returns a CSS color value string for a pixel cell.
//
// When obfuscate is false it returns a plain lowercase hex string (e.g.
// "#ff0000ff") suitable for use inside a background-color CSS property.
//
// When obfuscate is true it randomly picks from several CSS color notations
// — hex (lower/upper/mixed-case nibbles), rgba(), and hsla() — and also
// randomises the casing of the background-color property name. Since we use
// text/template, none of these formats are sanitised or blocked by the renderer.
func formatColor(r, g, b, a uint8, obfuscate bool) string {
	if !obfuscate {
		if a == 255 {
			return fmt.Sprintf("background-color:#%02x%02x%02x", r, g, b)
		}
		return fmt.Sprintf("background-color:#%02x%02x%02x%02x", r, g, b, a)
	}

	// Randomise the CSS color value representation.
	var colorVal string
	af := float64(a) / 255.0

	switch randIntn(5) {
	case 0:
		// Hex lowercase  e.g. #ff0000ff
		colorVal = fmt.Sprintf("#%02x%02x%02x%02x", r, g, b, a)
	case 1:
		// Hex uppercase  e.g. #FF0000FF
		colorVal = fmt.Sprintf("#%02X%02X%02X%02X", r, g, b, a)
	case 2:
		// Hex mixed-case per nibble  e.g. #fF00AaFF
		colorVal = fmt.Sprintf("#%s%s%s%s", hexByte(r), hexByte(g), hexByte(b), hexByte(a))
	case 3:
		// rgba() decimal  e.g. rgba(255,0,0,1.0)
		colorVal = fmt.Sprintf("rgba(%d,%d,%d,%.3g)", r, g, b, af)
	default:
		// hsla()  e.g. hsla(0,100%,50%,1.0)
		h, s, l := rgbToHSL(r, g, b)
		colorVal = fmt.Sprintf("hsla(%d,%d%%,%d%%,%.3g)", h, s, l, af)
	}

	// Randomise the CSS property name casing — background-color is
	prop := randomizeCase("background-color")

	return fmt.Sprintf("%s:%s", prop, colorVal)
}

// hexByte encodes a single byte as a two-character hex string with independently
// randomised case for each nibble, e.g. 0xff can yield "fF", "Ff", "ff", or "FF".
func hexByte(v uint8) string {
	const digits = "0123456789abcdef0123456789ABCDEF"
	hi := v >> 4
	lo := v & 0x0f
	hiOffset := uint8(randIntn(2)) * 16
	loOffset := uint8(randIntn(2)) * 16
	return string([]byte{digits[hiOffset+hi], digits[loOffset+lo]})
}

// randomizeCase returns s with each character randomly upper- or lower-cased.
// CSS property names are case-insensitive, so the visual result is identical.
func randomizeCase(s string) string {
	var sb strings.Builder
	sb.Grow(len(s))
	for _, c := range s {
		if randIntn(2) == 0 {
			sb.WriteString(strings.ToUpper(string(c)))
		} else {
			sb.WriteString(strings.ToLower(string(c)))
		}
	}
	return sb.String()
}

// rgbToHSL converts 8-bit RGB values to HSL (H: 0-360, S: 0-100, L: 0-100).
func rgbToHSL(r, g, b uint8) (int, int, int) {
	rf, gf, bf := float64(r)/255, float64(g)/255, float64(b)/255

	maxC := math.Max(rf, math.Max(gf, bf))
	minC := math.Min(rf, math.Min(gf, bf))
	l := (maxC + minC) / 2
	h, s := 0.0, 0.0

	if d := maxC - minC; d != 0 {
		if l > 0.5 {
			s = d / (2.0 - maxC - minC)
		} else {
			s = d / (maxC + minC)
		}
		switch maxC {
		case rf:
			h = (gf - bf) / d
			if gf < bf {
				h += 6.0
			}
		case gf:
			h = (bf-rf)/d + 2.0
		default: // bf
			h = (rf-gf)/d + 4.0
		}
		h /= 6.0
	}

	return int(h * 360), int(s * 100), int(l * 100)
}
