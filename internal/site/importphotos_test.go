package site

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf16"
)

func TestImportPhotos(t *testing.T) {
	root := t.TempDir()
	post, from := filepath.Join(root, "post"), filepath.Join(root, "from")
	if err := writeFile(filepath.Join(post, "index.md"), []byte("hi")); err != nil {
		t.Fatal(err)
	}
	writeJPEG(t, filepath.Join(from, "IMG_0001.JPG"), 3000, 1500)
	writeJPEG(t, filepath.Join(from, "small.jpeg"), 400, 300)
	var pngData bytes.Buffer
	if err := png.Encode(&pngData, image.NewNRGBA(image.Rect(0, 0, 10, 10))); err != nil {
		t.Fatal(err)
	}
	writeFile(filepath.Join(from, "Logo.png"), pngData.Bytes())
	writeFile(filepath.Join(from, "notes.txt"), []byte("not a photo"))

	done, err := ImportPhotos(post, from, false)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string][2]int{}
	for _, p := range done {
		got[filepath.Base(p.Path)] = [2]int{p.Width, p.Height}
	}
	want := map[string][2]int{
		"img_0001.jpg": {2560, 1280}, // scaled down to fit 2560
		"small.jpg":    {400, 300},   // never scaled up
		"logo.png":     {10, 10},     // PNG stays PNG
	}
	for name, size := range want {
		if got[name] != size {
			t.Errorf("%s: got %v, want %v (all: %v)", name, got[name], size, got)
		}
	}
	if len(got) != len(want) {
		t.Errorf("imported %v, want only the photos", got)
	}
	cfg, err := decodeConfig(filepath.Join(post, "img_0001.jpg"))
	if err != nil || cfg.Width != 2560 {
		t.Errorf("img_0001.jpg on disk is %dx%d (%v)", cfg.Width, cfg.Height, err)
	}

	// Importing again stops before writing anything, unless told to replace.
	if _, err := ImportPhotos(post, from, false); err == nil || !strings.Contains(err.Error(), "-replace") {
		t.Errorf("err = %v, want one that says to pass -replace", err)
	}
	if _, err := ImportPhotos(post, from, true); err != nil {
		t.Errorf("with replace: %v", err)
	}

	// A single photo works too.
	if done, err := ImportPhotos(post, filepath.Join(from, "small.jpeg"), true); err != nil || len(done) != 1 {
		t.Errorf("single photo: %v, %v", done, err)
	}

	if _, err := ImportPhotos(from, from, false); err == nil || !strings.Contains(err.Error(), "index.md") {
		t.Errorf("err = %v, want one saying it isn't a post folder", err)
	}
	writeFile(filepath.Join(from, "IMG_0002.HEIC"), []byte("heic"))
	if _, err := ImportPhotos(post, from, true); err == nil || !strings.Contains(err.Error(), "HEIC") {
		t.Errorf("err = %v, want one about HEIC", err)
	}
}

// jpegWithProfile wraps an ICC profile whose description is desc in a JPEG's
// APP2 segment, as a version 2 ('desc') or version 4 ('mluc') tag.
func jpegWithProfile(desc string, v4 bool) []byte {
	var tag []byte
	if v4 {
		units := utf16.Encode([]rune(desc))
		tag = append([]byte("mluc"), make([]byte, 24)...)
		binary.BigEndian.PutUint32(tag[8:], 1)
		binary.BigEndian.PutUint32(tag[12:], 12)
		copy(tag[16:], "enUS")
		binary.BigEndian.PutUint32(tag[20:], uint32(len(units)*2))
		binary.BigEndian.PutUint32(tag[24:], 28)
		for _, u := range units {
			tag = binary.BigEndian.AppendUint16(tag, u)
		}
	} else {
		tag = append([]byte("desc"), make([]byte, 8)...)
		binary.BigEndian.PutUint32(tag[8:], uint32(len(desc)+1))
		tag = append(append(tag, desc...), 0)
	}
	profile := make([]byte, 144)
	binary.BigEndian.PutUint32(profile[128:], 1)
	copy(profile[132:], "desc")
	binary.BigEndian.PutUint32(profile[136:], 144)
	binary.BigEndian.PutUint32(profile[140:], uint32(len(tag)))
	profile = append(profile, tag...)

	seg := append([]byte("ICC_PROFILE\x00\x01\x01"), profile...)
	out := []byte{0xFF, 0xD8, 0xFF, 0xE2}
	out = binary.BigEndian.AppendUint16(out, uint16(len(seg)+2))
	return append(append(out, seg...), 0xFF, 0xDA)
}

func TestJPEGICCName(t *testing.T) {
	if got := jpegICCName(jpegWithProfile("Display P3", true)); got != "Display P3" {
		t.Errorf("v4 profile: got %q", got)
	}
	if got := jpegICCName(jpegWithProfile("sRGB IEC61966-2.1", false)); got != "sRGB IEC61966-2.1" {
		t.Errorf("v2 profile: got %q", got)
	}
	if got := jpegICCName([]byte{0xFF, 0xD8, 0xFF, 0xDA}); got != "" {
		t.Errorf("no profile: got %q", got)
	}
	if got := jpegICCName([]byte("not a jpeg")); got != "" {
		t.Errorf("not a JPEG: got %q", got)
	}
}

func TestP3ToSRGB(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 2, 1))
	copy(img.Pix, []uint8{
		128, 128, 128, 255, // gray: the same in both
		0, 200, 0, 255, // P3 green: brighter in sRGB, and red and blue clip to 0
	})
	p3ToSRGB(img)
	want := []uint8{128, 128, 128, 255, 0, 204, 0, 255}
	if !bytes.Equal(img.Pix, want) {
		t.Errorf("got %v, want %v", img.Pix, want)
	}
}

func TestOpenPhotoRejectsOtherProfiles(t *testing.T) {
	path := filepath.Join(t.TempDir(), "adobe.jpg")
	writeJPEG(t, path, 10, 10)
	data, _ := os.ReadFile(path)
	// Splice an Adobe RGB profile in after the start-of-image marker.
	withProfile := jpegWithProfile("Adobe RGB (1998)", false)
	app2 := withProfile[2 : len(withProfile)-2]
	data = append(append([]byte{0xFF, 0xD8}, app2...), data[2:]...)
	os.WriteFile(path, data, 0o644)
	if _, err := openPhoto(path); err == nil || !strings.Contains(err.Error(), "Adobe RGB") {
		t.Errorf("err = %v, want one naming the profile", err)
	}
}
