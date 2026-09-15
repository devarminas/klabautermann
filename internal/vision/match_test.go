package vision

import (
	"image"
	"image/color"
	"testing"
)

func grayFrame(w, h int, bg uint8) *image.Gray {
	img := image.NewGray(image.Rect(0, 0, w, h))
	for i := range img.Pix {
		img.Pix[i] = bg
	}
	return img
}

func fillRect(img *image.Gray, r image.Rectangle, v uint8) {
	c := r.Intersect(img.Bounds())
	for y := c.Min.Y; y < c.Max.Y; y++ {
		for x := c.Min.X; x < c.Max.X; x++ {
			img.SetGray(x, y, color.Gray{Y: v})
		}
	}
}

func drawRect(img *image.Gray, r image.Rectangle, border, inner uint8) {
	fillRect(img, r, border)
	fillRect(img, image.Rect(r.Min.X+2, r.Min.Y+2, r.Max.X-2, r.Max.Y-2), inner)
}

func cropGray(src *image.Gray, r image.Rectangle) *image.Gray {
	dst := image.NewGray(image.Rect(0, 0, r.Dx(), r.Dy()))
	for y := 0; y < r.Dy(); y++ {
		for x := 0; x < r.Dx(); x++ {
			dst.SetGray(x, y, src.GrayAt(r.Min.X+x, r.Min.Y+y))
		}
	}
	return dst
}

func TestMatchSameScreenHit(t *testing.T) {
	frame := grayFrame(200, 150, 255)
	src := image.Rect(50, 60, 90, 80)
	drawRect(frame, src, 0, 255)
	tmpl := cropGray(frame, src)
	got, err := Match(frame, tmpl, frame.Bounds())
	if err != nil {
		t.Fatalf("Match returned error: %v", err)
	}
	if got.Rect != src {
		t.Fatalf("Match rect = %v, want %v", got.Rect, src)
	}
	if (got.Center != image.Point{X: 70, Y: 70}) {
		t.Fatalf("Match center = %v, want %v", got.Center, image.Point{X: 70, Y: 70})
	}
	if got.Score <= 0.99 {
		t.Fatalf("Match score = %v, want above 0.99", got.Score)
	}
}

func TestMatchMiss(t *testing.T) {
	with := grayFrame(200, 150, 255)
	drawRect(with, image.Rect(50, 60, 90, 80), 0, 255)
	tmpl := cropGray(with, image.Rect(50, 60, 90, 80))
	frame := grayFrame(200, 150, 255)
	got, err := Match(frame, tmpl, frame.Bounds())
	if err != nil {
		t.Fatalf("Match returned error: %v", err)
	}
	if got.Score >= 0.8 {
		t.Fatalf("Match score = %v, want below 0.8", got.Score)
	}
}

func TestMatchShift(t *testing.T) {
	with := grayFrame(200, 150, 255)
	drawRect(with, image.Rect(50, 60, 90, 80), 0, 255)
	tmpl := cropGray(with, image.Rect(50, 60, 90, 80))
	frame := grayFrame(200, 150, 255)
	want := image.Rect(60, 60, 100, 80)
	drawRect(frame, want, 0, 255)
	got, err := Match(frame, tmpl, image.Rect(40, 50, 120, 90))
	if err != nil {
		t.Fatalf("Match returned error: %v", err)
	}
	if got.Rect != want {
		t.Fatalf("Match rect = %v, want %v", got.Rect, want)
	}
	if (got.Center != image.Point{X: 80, Y: 70}) {
		t.Fatalf("Match center = %v, want %v", got.Center, image.Point{X: 80, Y: 70})
	}
	if got.Score <= 0.99 {
		t.Fatalf("Match score = %v, want above 0.99", got.Score)
	}
}

func TestMatchBrightnessInvariance(t *testing.T) {
	with := grayFrame(200, 150, 200)
	src := image.Rect(50, 60, 90, 80)
	drawRect(with, src, 100, 200)
	tmpl := cropGray(with, src)
	frame := grayFrame(200, 150, 220)
	drawRect(frame, src, 120, 220)
	got, err := Match(frame, tmpl, frame.Bounds())
	if err != nil {
		t.Fatalf("Match returned error: %v", err)
	}
	if got.Rect != src {
		t.Fatalf("Match rect = %v, want %v", got.Rect, src)
	}
	if (got.Center != image.Point{X: 70, Y: 70}) {
		t.Fatalf("Match center = %v, want %v", got.Center, image.Point{X: 70, Y: 70})
	}
	if got.Score <= 0.99 {
		t.Fatalf("Match score = %v, want above 0.99", got.Score)
	}
}

func TestMatchErrors(t *testing.T) {
	frame := grayFrame(200, 150, 255)
	drawRect(frame, image.Rect(50, 60, 90, 80), 0, 255)
	tmpl := cropGray(frame, image.Rect(50, 60, 90, 80))
	if _, err := Match(frame, tmpl, image.Rect(0, 0, 10, 10)); err == nil {
		t.Fatalf("template larger than ROI: expected error, got nil")
	}
	if _, err := Match(frame, image.NewGray(image.Rectangle{}), frame.Bounds()); err == nil {
		t.Fatalf("empty template: expected error, got nil")
	}
	if _, err := Match(frame, tmpl, image.Rect(500, 500, 600, 600)); err == nil {
		t.Fatalf("ROI outside frame: expected error, got nil")
	}
}
