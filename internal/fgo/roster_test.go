package fgo_test

import (
	"image"
	"image/color"
	"math"
	"testing"

	"klabautermann/internal/fgo"
)

func checker(w, h int, a, b color.Color) *image.RGBA {
	m := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if (x+y)%2 == 0 {
				m.Set(x, y, a)
			} else {
				m.Set(x, y, b)
			}
		}
	}
	return m
}

func paste(dst *image.RGBA, src image.Image, ox, oy int) {
	sb := src.Bounds()
	for y := sb.Min.Y; y < sb.Max.Y; y++ {
		for x := sb.Min.X; x < sb.Max.X; x++ {
			dst.Set(ox+x-sb.Min.X, oy+y-sb.Min.Y, src.At(x, y))
		}
	}
}

func TestDetectClasses(t *testing.T) {
	icon := checker(20, 20, color.Black, color.White)
	frame := image.NewRGBA(image.Rect(0, 0, 200, 200))
	gray := color.RGBA{128, 128, 128, 255}
	for y := 0; y < 200; y++ {
		for x := 0; x < 200; x++ {
			frame.Set(x, y, gray)
		}
	}
	ox, oy := 50, 70
	paste(frame, icon, ox, oy)
	want := image.Rect(ox, oy, ox+20, oy+20)
	icons := map[fgo.Class]image.Image{fgo.Saber: icon}
	got := fgo.DetectClasses(frame, icons, want, 0.8)
	if len(got) != 1 {
		t.Fatalf("got %d detections, want 1", len(got))
	}
	if got[0].Class != fgo.Saber {
		t.Fatalf("got class %q, want %q", got[0].Class, fgo.Saber)
	}
	if math.Abs(got[0].Score-1.0) > 1e-9 {
		t.Fatalf("got score %v, want 1.0", got[0].Score)
	}
	if got[0].At != want {
		t.Fatalf("got rect %v, want %v", got[0].At, want)
	}
	other := checker(20, 20, color.Black, color.RGBA{0, 0, 0, 255})
	for y := 0; y < 20; y++ {
		for x := 10; x < 20; x++ {
			other.Set(x, y, color.White)
		}
	}
	mixed := map[fgo.Class]image.Image{fgo.Saber: icon, fgo.Archer: other}
	mixedGot := fgo.DetectClasses(frame, mixed, want, 0.8)
	for _, d := range mixedGot {
		if d.Class == fgo.Archer {
			t.Fatalf("different checker excluded, got %+v", d)
		}
	}
	if len(mixedGot) != 1 || mixedGot[0].Class != fgo.Saber {
		t.Fatalf("got %+v, want single Saber detection", mixedGot)
	}
	full := fgo.DetectClasses(frame, icons, image.Rectangle{}, 0.8)
	if len(full) != 1 || full[0].At != want {
		t.Fatalf("empty roi got %+v, want rect %v", full, want)
	}
}

func TestGaugeFill(t *testing.T) {
	dark := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			dark.Set(x, y, color.Black)
		}
	}
	if v := fgo.GaugeFill(dark, dark.Bounds()); v != 0.0 {
		t.Fatalf("dark got %v, want 0.0", v)
	}
	bright := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			bright.Set(x, y, color.White)
		}
	}
	if v := fgo.GaugeFill(bright, bright.Bounds()); v != 1.0 {
		t.Fatalf("bright got %v, want 1.0", v)
	}
	bar := image.NewRGBA(image.Rect(0, 0, 100, 1))
	for x := 0; x < 100; x++ {
		var c color.Color = color.Black
		if x < 37 {
			c = color.White
		}
		bar.Set(x, 0, c)
	}
	if v := fgo.GaugeFill(bar, bar.Bounds()); math.Abs(v-0.37) > 0.02 {
		t.Fatalf("bar got %v, want 0.37", v)
	}
	if v := fgo.GaugeFill(bar, image.Rectangle{}); v != 0 {
		t.Fatalf("empty rect got %v, want 0", v)
	}
}
