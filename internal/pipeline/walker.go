package pipeline

import (
	"context"
	"fmt"
	"image"
	"time"

	"klabautermann/internal/vision"
)

type Node struct {
	Name      string
	Template  image.Image
	ROI       image.Rectangle
	Threshold float64
	TapCenter bool
	TapPoint  image.Point
	Next      []string
	OnError   string
	Scan      *ScanSpec
}

// ScanSpec is one drag gesture plus a swipe budget.
type ScanSpec struct {
	X1, Y1, X2, Y2 int
	DurationMs     int
	MaxSwipes      int
}

type Hit struct {
	Node  string
	Score float64
	At    image.Rectangle
}

type Walker struct {
	Nodes     map[string]Node
	Capture   func(ctx context.Context) (image.Image, error)
	Tap       func(ctx context.Context, x, y int) error
	Swipe     func(ctx context.Context, x1, y1, x2, y2, ms int) error
	MaxMisses int
	MaxSteps  int
	Settle    time.Duration
}

func (w *Walker) Run(ctx context.Context, start string) ([]Hit, error) {
	nodes := make(map[string]Node, len(w.Nodes))
	for name, n := range w.Nodes {
		if n.Threshold <= 0 {
			n.Threshold = 0.8
		}
		nodes[name] = n
	}
	if _, ok := nodes[start]; !ok {
		return nil, fmt.Errorf("pipeline: unknown start node %q", start)
	}
	for name, n := range nodes {
		if n.Threshold > 1 {
			return nil, fmt.Errorf("pipeline: node %q threshold %v outside (0, 1]", name, n.Threshold)
		}
		for _, next := range n.Next {
			if _, ok := nodes[next]; !ok {
				return nil, fmt.Errorf("pipeline: node %q next %q names unknown node", name, next)
			}
		}
		if n.OnError != "" {
			if _, ok := nodes[n.OnError]; !ok {
				return nil, fmt.Errorf("pipeline: node %q onerror %q names unknown node", name, n.OnError)
			}
		}
	}
	maxMisses := w.MaxMisses
	if maxMisses <= 0 {
		maxMisses = 3
	}
	maxSteps := w.MaxSteps
	if maxSteps <= 0 {
		maxSteps = 100
	}
	var hits []Hit
	candidates := []string{start}
	owner := start
	misses := 0
	steps := 0
	swipes := make(map[string]int)
	for {
		if err := ctx.Err(); err != nil {
			return hits, fmt.Errorf("pipeline: context: %w", err)
		}
		frame, err := w.Capture(ctx)
		if err != nil {
			return hits, fmt.Errorf("pipeline: capture: %w", err)
		}
		var current Node
		var currentName string
		var currentRes vision.MatchResult
		found := false
		for _, name := range candidates {
			n := nodes[name]
			if n.ROI.Empty() {
				n.ROI = frame.Bounds()
			}
			res, err := vision.Match(frame, n.Template, n.ROI)
			if err != nil {
				continue
			}
			if res.Score < n.Threshold {
				continue
			}
			current = n
			currentName = name
			currentRes = res
			found = true
			break
		}
		if !found {
			o := nodes[owner]
			if o.Scan != nil {
				budget := o.Scan.MaxSwipes
				if budget <= 0 {
					budget = 5
				}
				if swipes[owner] < budget {
					if w.Swipe == nil {
						return hits, fmt.Errorf("pipeline: node %q has scan but swipe unsupported", owner)
					}
					if err := w.Swipe(ctx, o.Scan.X1, o.Scan.Y1, o.Scan.X2, o.Scan.Y2, o.Scan.DurationMs); err != nil {
						return hits, fmt.Errorf("pipeline: swipe node %q: %w", owner, err)
					}
					if w.Settle > 0 {
						select {
						case <-ctx.Done():
							return hits, fmt.Errorf("pipeline: context: %w", ctx.Err())
						case <-time.After(w.Settle):
						}
					}
					swipes[owner]++
					continue
				}
			}
			misses++
			if misses <= maxMisses {
				continue
			}
			if o.OnError == "" {
				return hits, fmt.Errorf("pipeline: node %q missed %d frames with no recovery", owner, misses)
			}
			candidates = []string{o.OnError}
			owner = o.OnError
			misses = 0
			continue
		}
		hits = append(hits, Hit{Node: currentName, Score: currentRes.Score, At: currentRes.Rect})
		if current.TapCenter {
			err = w.Tap(ctx, currentRes.Center.X, currentRes.Center.Y)
		} else {
			err = w.Tap(ctx, current.TapPoint.X, current.TapPoint.Y)
		}
		if err != nil {
			return hits, fmt.Errorf("pipeline: tap node %q: %w", currentName, err)
		}
		if w.Settle > 0 {
			select {
			case <-ctx.Done():
				return hits, fmt.Errorf("pipeline: context: %w", ctx.Err())
			case <-time.After(w.Settle):
			}
		}
		steps++
		if steps > maxSteps {
			return hits, fmt.Errorf("pipeline: exceeded max steps %d", maxSteps)
		}
		if len(current.Next) == 0 {
			return hits, nil
		}
		candidates = append([]string(nil), current.Next...)
		owner = currentName
		misses = 0
		swipes[currentName] = 0
	}
}
