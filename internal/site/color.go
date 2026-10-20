package site

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"math"
	"os"
	"strings"
	"unicode/utf16"

	"github.com/disintegration/imaging"
)

// openPhoto decodes a photo, upright and in sRGB.
//
// Phone cameras store the rotation in EXIF rather than turning the pixels, so
// that's applied. Phones also often shoot in Display P3, a wider color space
// than sRGB, and say so in an embedded ICC profile. Go's decoders ignore the
// profile, and the files written from here don't carry one, so browsers would
// read the P3 values as sRGB and show them muted. Converting to sRGB keeps
// the colors as they were.
func openPhoto(path string) (*image.NRGBA, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	decoded, err := imaging.Decode(bytes.NewReader(data), imaging.AutoOrientation(true))
	if err != nil {
		return nil, err
	}
	img := imaging.Clone(decoded)

	switch name := jpegICCName(data); {
	case name == "" || strings.Contains(name, "sRGB"):
		return img, nil
	case name == "Display P3":
		p3ToSRGB(img)
		return img, nil
	default:
		return nil, fmt.Errorf("its color profile is %q; only sRGB and Display P3 are supported, so export it as sRGB", name)
	}
}

// jpegICCName returns the description of the ICC profile embedded in a JPEG,
// like "Display P3", or "" when there isn't one (or it isn't a JPEG).
func jpegICCName(data []byte) string {
	if len(data) < 4 || data[0] != 0xFF || data[1] != 0xD8 {
		return ""
	}
	// The profile is split across APP2 segments, each starting with
	// "ICC_PROFILE\0", a sequence number and a count.
	var profile []byte
	for i := 2; i+4 <= len(data) && data[i] == 0xFF; {
		marker := data[i+1]
		if marker == 0xDA { // start of scan: no more metadata
			break
		}
		size := int(binary.BigEndian.Uint16(data[i+2:]))
		end := i + 2 + size
		if size < 2 || end > len(data) {
			break
		}
		seg := data[i+4 : end]
		if marker == 0xE2 && bytes.HasPrefix(seg, []byte("ICC_PROFILE\x00")) && len(seg) >= 14 {
			profile = append(profile, seg[14:]...)
		}
		i = end
	}
	return iccDescription(profile)
}

// iccDescription reads the 'desc' tag of an ICC profile: ASCII in version 2
// profiles, or UTF-16 in version 4's 'mluc' type.
func iccDescription(p []byte) string {
	if len(p) < 132 {
		return ""
	}
	count := int(binary.BigEndian.Uint32(p[128:]))
	for i := range count {
		entry := 132 + i*12
		if entry+12 > len(p) {
			return ""
		}
		if string(p[entry:entry+4]) != "desc" {
			continue
		}
		off := int(binary.BigEndian.Uint32(p[entry+4:]))
		size := int(binary.BigEndian.Uint32(p[entry+8:]))
		if off+size > len(p) || size < 12 {
			return ""
		}
		tag := p[off : off+size]
		switch string(tag[:4]) {
		case "desc":
			n := int(binary.BigEndian.Uint32(tag[8:]))
			if 12+n > len(tag) {
				return ""
			}
			return strings.TrimRight(string(tag[12:12+n]), "\x00")
		case "mluc":
			if len(tag) < 28 {
				return ""
			}
			n := int(binary.BigEndian.Uint32(tag[20:]))
			start := int(binary.BigEndian.Uint32(tag[24:]))
			if start+n > len(tag) {
				return ""
			}
			units := make([]uint16, n/2)
			for j := range units {
				units[j] = binary.BigEndian.Uint16(tag[start+j*2:])
			}
			return string(utf16.Decode(units))
		}
		return ""
	}
	return ""
}

// p3ToSRGB converts img's pixels from Display P3 to sRGB, in place. Both use
// the same transfer curve and white point, so it's: undo the curve, apply the
// matrix between the two sets of primaries, clip what falls outside sRGB, and
// put the curve back.
func p3ToSRGB(img *image.NRGBA) {
	m := [3][3]float64{
		{1.2249401, -0.2249404, 0},
		{-0.0420569, 1.0420571, 0},
		{-0.0196376, -0.0786361, 1.0982735},
	}
	for i := 0; i+3 < len(img.Pix); i += 4 {
		r, g, b := toLinear[img.Pix[i]], toLinear[img.Pix[i+1]], toLinear[img.Pix[i+2]]
		img.Pix[i] = fromLinear(m[0][0]*r + m[0][1]*g + m[0][2]*b)
		img.Pix[i+1] = fromLinear(m[1][0]*r + m[1][1]*g + m[1][2]*b)
		img.Pix[i+2] = fromLinear(m[2][0]*r + m[2][1]*g + m[2][2]*b)
	}
}

// toLinear undoes the sRGB (and Display P3) transfer curve for each 8-bit value.
var toLinear = func() (t [256]float64) {
	for i := range t {
		v := float64(i) / 255
		if v <= 0.04045 {
			t[i] = v / 12.92
		} else {
			t[i] = math.Pow((v+0.055)/1.055, 2.4)
		}
	}
	return t
}()

// linearSteps is how finely fromLinearTable samples the curve; 4096 steps keep
// dark tones, where the curve is steepest, from banding.
const linearSteps = 4096

var fromLinearTable = func() (t [linearSteps + 1]uint8) {
	for i := range t {
		v := float64(i) / linearSteps
		if v <= 0.0031308 {
			v *= 12.92
		} else {
			v = 1.055*math.Pow(v, 1/2.4) - 0.055
		}
		t[i] = uint8(math.Round(v * 255))
	}
	return t
}()

// fromLinear puts the transfer curve back on a linear value, clipped to 0–1.
func fromLinear(v float64) uint8 {
	return fromLinearTable[int(math.Round(min(max(v, 0), 1)*linearSteps))]
}
