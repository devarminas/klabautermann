package pipeline

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
)

type Pipeline struct {
	Start string
	Nodes map[string]Node
}

type rawPoint struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type rawRect struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

type rawNode struct {
	Name      string          `json:"name"`
	Template  string          `json:"template"`
	Threshold float64         `json:"threshold"`
	Tap       json.RawMessage `json:"tap"`
	ROI       *rawRect        `json:"roi"`
	Next      []string        `json:"next"`
	OnError   string          `json:"on_error"`
}

type rawPipeline struct {
	Start string    `json:"start"`
	Nodes []rawNode `json:"nodes"`
}

func Load(path string) (Pipeline, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Pipeline{}, fmt.Errorf("pipeline: load %s: %w", path, err)
	}
	var raw rawPipeline
	if err := json.Unmarshal(data, &raw); err != nil {
		return Pipeline{}, fmt.Errorf("pipeline: load %s: invalid JSON: %w", path, err)
	}
	if raw.Start == "" {
		return Pipeline{}, fmt.Errorf("pipeline: load %s: missing start", path)
	}
	dir := filepath.Dir(path)
	nodes := make(map[string]Node, len(raw.Nodes))
	for i := range raw.Nodes {
		rn := raw.Nodes[i]
		if rn.Name == "" {
			return Pipeline{}, fmt.Errorf("pipeline: load %s: node %d missing name", path, i)
		}
		if _, dup := nodes[rn.Name]; dup {
			return Pipeline{}, fmt.Errorf("pipeline: load %s: duplicate node %q", path, rn.Name)
		}
		if rn.Template == "" {
			return Pipeline{}, fmt.Errorf("pipeline: load %s: node %q missing template", path, rn.Name)
		}
		tmplData, err := os.ReadFile(filepath.Join(dir, rn.Template))
		if err != nil {
			return Pipeline{}, fmt.Errorf("pipeline: load %s: node %q template %q: %w", path, rn.Name, rn.Template, err)
		}
		img, _, err := image.Decode(bytes.NewReader(tmplData))
		if err != nil {
			return Pipeline{}, fmt.Errorf("pipeline: load %s: node %q template %q: %w", path, rn.Name, rn.Template, err)
		}
		n := Node{
			Name:      rn.Name,
			Template:  img,
			Threshold: rn.Threshold,
			TapCenter: true,
			Next:      rn.Next,
			OnError:   rn.OnError,
		}
		if len(rn.Tap) > 0 && string(rn.Tap) != "null" {
			var s string
			if err := json.Unmarshal(rn.Tap, &s); err == nil {
				if s != "center" {
					return Pipeline{}, fmt.Errorf("pipeline: load %s: node %q tap %q unknown", path, rn.Name, s)
				}
			} else {
				var p rawPoint
				if err := json.Unmarshal(rn.Tap, &p); err != nil {
					return Pipeline{}, fmt.Errorf("pipeline: load %s: node %q tap invalid: %w", path, rn.Name, err)
				}
				n.TapCenter = false
				n.TapPoint = image.Point{X: p.X, Y: p.Y}
			}
		}
		if rn.ROI != nil {
			n.ROI = image.Rect(rn.ROI.X, rn.ROI.Y, rn.ROI.X+rn.ROI.W, rn.ROI.Y+rn.ROI.H)
		}
		nodes[rn.Name] = n
	}
	if _, ok := nodes[raw.Start]; !ok {
		return Pipeline{}, fmt.Errorf("pipeline: load %s: start %q names unknown node", path, raw.Start)
	}
	return Pipeline{Start: raw.Start, Nodes: nodes}, nil
}
