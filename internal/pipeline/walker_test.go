package pipeline

import (
	"context"
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

func drawMarker(img *image.Gray, r image.Rectangle) {
	fillRect(img, r, 0)
	fillRect(img, image.Rect(r.Min.X+2, r.Min.Y+2, r.Max.X-2, r.Max.Y-2), 255)
}

func markerFrame(rects ...image.Rectangle) *image.Gray {
	f := grayFrame(200, 150, 255)
	for _, r := range rects {
		drawMarker(f, r)
	}
	return f
}

func templateFor(r image.Rectangle) image.Image {
	src := markerFrame(r)
	dst := image.NewGray(image.Rect(0, 0, r.Dx(), r.Dy()))
	for y := 0; y < r.Dy(); y++ {
		for x := 0; x < r.Dx(); x++ {
			dst.SetGray(x, y, src.GrayAt(r.Min.X+x, r.Min.Y+y))
		}
	}
	return dst
}

func frameBounds() image.Rectangle {
	return image.Rect(0, 0, 200, 150)
}

func tapNode(name string, r image.Rectangle, next ...string) Node {
	return Node{Name: name, Template: templateFor(r), ROI: frameBounds(), Threshold: 0.8, TapCenter: true, Next: next}
}

func scripted(frames ...image.Image) func(ctx context.Context) (image.Image, error) {
	i := 0
	return func(ctx context.Context) (image.Image, error) {
		f := frames[i]
		if i < len(frames)-1 {
			i++
		}
		return f, nil
	}
}

func hitNames(hits []Hit) []string {
	names := make([]string, len(hits))
	for i, h := range hits {
		names[i] = h.Node
	}
	return names
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func equalPoints(a, b []image.Point) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestWalk(t *testing.T) {
	home := image.Rect(20, 30, 60, 50)
	quest := image.Rect(100, 40, 130, 66)
	noap := image.Rect(140, 90, 164, 120)
	w := &Walker{
		Nodes: map[string]Node{
			"Home":  tapNode("Home", home, "Quest"),
			"Quest": tapNode("Quest", quest, "NoAP"),
			"NoAP":  tapNode("NoAP", noap),
		},
		Capture: scripted(markerFrame(home), markerFrame(quest), markerFrame(noap)),
	}
	var taps []image.Point
	w.Tap = func(ctx context.Context, x, y int) error {
		taps = append(taps, image.Point{X: x, Y: y})
		return nil
	}
	hits, err := w.Run(context.Background(), "Home")
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if !equalStrings(hitNames(hits), []string{"Home", "Quest", "NoAP"}) {
		t.Fatalf("hits = %v, want [Home Quest NoAP]", hitNames(hits))
	}
	for i, r := range []image.Rectangle{home, quest, noap} {
		if hits[i].At != r {
			t.Fatalf("hits[%d].At = %v, want %v", i, hits[i].At, r)
		}
	}
	wantTaps := []image.Point{{X: 40, Y: 40}, {X: 115, Y: 53}, {X: 152, Y: 105}}
	if !equalPoints(taps, wantTaps) {
		t.Fatalf("taps = %v, want %v", taps, wantTaps)
	}
}

func TestFirstMatchWins(t *testing.T) {
	s := image.Rect(10, 10, 54, 26)
	a := image.Rect(70, 30, 110, 50)
	b := image.Rect(130, 80, 150, 116)
	w := &Walker{
		Nodes: map[string]Node{
			"S": tapNode("S", s, "A", "B"),
			"A": tapNode("A", a),
			"B": tapNode("B", b),
		},
		Capture: scripted(markerFrame(s), markerFrame(a, b)),
		Tap:     func(ctx context.Context, x, y int) error { return nil },
	}
	hits, err := w.Run(context.Background(), "S")
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if !equalStrings(hitNames(hits), []string{"S", "A"}) {
		t.Fatalf("hits = %v, want [S A]", hitNames(hits))
	}
}

func TestSecondCandidate(t *testing.T) {
	s := image.Rect(10, 10, 54, 26)
	a := image.Rect(70, 30, 110, 50)
	b := image.Rect(130, 80, 150, 116)
	w := &Walker{
		Nodes: map[string]Node{
			"S": tapNode("S", s, "A", "B"),
			"A": tapNode("A", a),
			"B": tapNode("B", b),
		},
		Capture: scripted(markerFrame(s), markerFrame(b)),
		Tap:     func(ctx context.Context, x, y int) error { return nil },
	}
	hits, err := w.Run(context.Background(), "S")
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if !equalStrings(hitNames(hits), []string{"S", "B"}) {
		t.Fatalf("hits = %v, want [S B]", hitNames(hits))
	}
	if hits[1].At != b {
		t.Fatalf("hits[1].At = %v, want %v", hits[1].At, b)
	}
}

func TestRecovery(t *testing.T) {
	s := image.Rect(10, 10, 54, 26)
	ghost := image.Rect(70, 30, 110, 50)
	rec := image.Rect(130, 80, 150, 116)
	start := tapNode("S", s, "Ghost")
	start.OnError = "R"
	w := &Walker{
		Nodes: map[string]Node{
			"S":     start,
			"Ghost": tapNode("Ghost", ghost),
			"R":     tapNode("R", rec),
		},
		Capture:   scripted(markerFrame(s), markerFrame(), markerFrame(), markerFrame(rec)),
		Tap:       func(ctx context.Context, x, y int) error { return nil },
		MaxMisses: 1,
	}
	hits, err := w.Run(context.Background(), "S")
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if !equalStrings(hitNames(hits), []string{"S", "R"}) {
		t.Fatalf("hits = %v, want [S R]", hitNames(hits))
	}
}

func TestExhaustion(t *testing.T) {
	s := image.Rect(10, 10, 54, 26)
	ghost := image.Rect(70, 30, 110, 50)
	w := &Walker{
		Nodes: map[string]Node{
			"S":     tapNode("S", s, "Ghost"),
			"Ghost": tapNode("Ghost", ghost),
		},
		Capture: scripted(markerFrame(s), markerFrame()),
		Tap:     func(ctx context.Context, x, y int) error { return nil },
	}
	hits, err := w.Run(context.Background(), "S")
	if err == nil {
		t.Fatalf("expected exhaustion error, got nil")
	}
	if !equalStrings(hitNames(hits), []string{"S"}) {
		t.Fatalf("hits = %v, want [S]", hitNames(hits))
	}
}

func TestBadGraph(t *testing.T) {
	s := image.Rect(10, 10, 54, 26)
	build := func(calls *int) *Walker {
		return &Walker{
			Nodes: map[string]Node{
				"Home": tapNode("Home", s, "Ghost"),
			},
			Capture: func(ctx context.Context) (image.Image, error) {
				*calls++
				return markerFrame(s), nil
			},
			Tap: func(ctx context.Context, x, y int) error { return nil },
		}
	}
	calls := 0
	if _, err := build(&calls).Run(context.Background(), "Nope"); err == nil {
		t.Fatalf("unknown start: expected error, got nil")
	}
	if calls != 0 {
		t.Fatalf("unknown start: capture calls = %d, want 0", calls)
	}
	calls = 0
	if _, err := build(&calls).Run(context.Background(), "Home"); err == nil {
		t.Fatalf("unknown next: expected error, got nil")
	}
	if calls != 0 {
		t.Fatalf("unknown next: capture calls = %d, want 0", calls)
	}
}

func TestStepBudget(t *testing.T) {
	a := image.Rect(70, 30, 110, 50)
	b := image.Rect(130, 80, 150, 116)
	w := &Walker{
		Nodes: map[string]Node{
			"A": tapNode("A", a, "B"),
			"B": tapNode("B", b, "A"),
		},
		Capture:  scripted(markerFrame(a, b)),
		Tap:      func(ctx context.Context, x, y int) error { return nil },
		MaxSteps: 3,
	}
	hits, err := w.Run(context.Background(), "A")
	if err == nil {
		t.Fatalf("expected step budget error, got nil")
	}
	if !equalStrings(hitNames(hits), []string{"A", "B", "A", "B"}) {
		t.Fatalf("hits = %v, want [A B A B]", hitNames(hits))
	}
}

func TestZeroROIUsesFullFrame(t *testing.T) {
	r := image.Rect(140, 90, 164, 120)
	n := tapNode("Z", r)
	n.ROI = image.Rectangle{}
	w := &Walker{
		Nodes:   map[string]Node{"Z": n},
		Capture: scripted(markerFrame(r)),
		Tap:     func(ctx context.Context, x, y int) error { return nil },
	}
	hits, err := w.Run(context.Background(), "Z")
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if !equalStrings(hitNames(hits), []string{"Z"}) {
		t.Fatalf("hits = %v, want [Z]", hitNames(hits))
	}
	if hits[0].At != r {
		t.Fatalf("hits[0].At = %v, want %v", hits[0].At, r)
	}
}

func TestFixedTap(t *testing.T) {
	r := image.Rect(70, 30, 110, 50)
	n := tapNode("F", r)
	n.TapCenter = false
	n.TapPoint = image.Point{X: 7, Y: 9}
	var taps []image.Point
	w := &Walker{
		Nodes:   map[string]Node{"F": n},
		Capture: scripted(markerFrame(r)),
		Tap: func(ctx context.Context, x, y int) error {
			taps = append(taps, image.Point{X: x, Y: y})
			return nil
		},
	}
	hits, err := w.Run(context.Background(), "F")
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if !equalStrings(hitNames(hits), []string{"F"}) {
		t.Fatalf("hits = %v, want [F]", hitNames(hits))
	}
	if !equalPoints(taps, []image.Point{{X: 7, Y: 9}}) {
		t.Fatalf("taps = %v, want [{7 9}]", taps)
	}
}
