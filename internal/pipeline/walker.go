package pipeline

import (
	"context"
	"fmt"
	"image"

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
	MaxMisses int
	MaxSteps  int
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
			misses++
			if misses <= maxMisses {
				continue
			}
			o := nodes[owner]
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
	}
}
