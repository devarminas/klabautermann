package pipeline

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func writeTestPNG(t *testing.T, path string, w, h int) {
	t.Helper()
	img := image.NewGray(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.SetGray(x, y, color.Gray{Y: uint8((x + y) % 2 * 255)})
		}
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
}

func writeTestPack(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	tmpl := filepath.Join(root, "tmpl")
	pack := filepath.Join(root, "pack")
	if err := os.Mkdir(tmpl, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(pack, 0755); err != nil {
		t.Fatal(err)
	}
	writeTestPNG(t, filepath.Join(tmpl, "home.png"), 4, 3)
	writeTestPNG(t, filepath.Join(tmpl, "formation.png"), 5, 6)
	writeTestPNG(t, filepath.Join(tmpl, "close.png"), 7, 8)
	doc := `{
  "start": "Home",
  "nodes": [
    {"name": "Home", "template": "../tmpl/home.png", "threshold": 0.8, "tap": "center", "roi": {"x": 1, "y": 2, "w": 10, "h": 8}, "next": ["Formation"]},
    {"name": "Formation", "template": "../tmpl/formation.png", "tap": {"x": 7, "y": 9}, "next": ["Close"], "on_error": "Close"},
    {"name": "Close", "template": "../tmpl/close.png", "next": []}
  ]
}`
	path := filepath.Join(pack, "pipe.json")
	if err := os.WriteFile(path, []byte(doc), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoad(t *testing.T) {
	p, err := Load(writeTestPack(t))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if p.Start != "Home" {
		t.Fatalf("Start = %q, want %q", p.Start, "Home")
	}
	if len(p.Nodes) != 3 {
		t.Fatalf("node count = %d, want 3", len(p.Nodes))
	}
	home := p.Nodes["Home"]
	if home.Template.Bounds() != image.Rect(0, 0, 4, 3) {
		t.Fatalf("Home template bounds = %v, want (0,0)-(4,3)", home.Template.Bounds())
	}
	if !home.TapCenter {
		t.Fatalf("Home TapCenter = false, want true")
	}
	if home.Threshold != 0.8 {
		t.Fatalf("Home Threshold = %v, want 0.8", home.Threshold)
	}
	if home.ROI != image.Rect(1, 2, 11, 10) {
		t.Fatalf("Home ROI = %v, want (1,2)-(11,10)", home.ROI)
	}
	formation := p.Nodes["Formation"]
	if formation.TapCenter {
		t.Fatalf("Formation TapCenter = true, want false")
	}
	if formation.TapPoint != (image.Point{X: 7, Y: 9}) {
		t.Fatalf("Formation TapPoint = %v, want (7,9)", formation.TapPoint)
	}
	if formation.Threshold != 0 {
		t.Fatalf("Formation Threshold = %v, want 0", formation.Threshold)
	}
	if !formation.ROI.Empty() {
		t.Fatalf("Formation ROI = %v, want empty", formation.ROI)
	}
	if formation.OnError != "Close" {
		t.Fatalf("Formation OnError = %q, want %q", formation.OnError, "Close")
	}
	close := p.Nodes["Close"]
	if !close.TapCenter {
		t.Fatalf("Close TapCenter = false, want true")
	}
	if close.Template.Bounds() != image.Rect(0, 0, 7, 8) {
		t.Fatalf("Close template bounds = %v, want (0,0)-(7,8)", close.Template.Bounds())
	}
}

func TestLoadErrors(t *testing.T) {
	if _, err := Load(filepath.Join(t.TempDir(), "missing.json")); err == nil {
		t.Fatalf("missing file: expected error, got nil")
	}
	dir := t.TempDir()
	badJSON := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(badJSON, []byte(`{"start": `), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(badJSON); err == nil {
		t.Fatalf("malformed JSON: expected error, got nil")
	}
	missingTmpl := filepath.Join(dir, "missing-tmpl.json")
	doc := `{"start": "A", "nodes": [{"name": "A", "template": "nope.png", "next": []}]}`
	if err := os.WriteFile(missingTmpl, []byte(doc), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(missingTmpl); err == nil {
		t.Fatalf("missing template: expected error, got nil")
	}
	corruptPath := filepath.Join(dir, "corrupt.png")
	if err := os.WriteFile(corruptPath, []byte("not image data"), 0644); err != nil {
		t.Fatal(err)
	}
	corruptDoc := filepath.Join(dir, "corrupt.json")
	doc = `{"start": "A", "nodes": [{"name": "A", "template": "corrupt.png", "next": []}]}`
	if err := os.WriteFile(corruptDoc, []byte(doc), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(corruptDoc); err == nil {
		t.Fatalf("corrupt template: expected error, got nil")
	}
}
