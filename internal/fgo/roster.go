package fgo

import (
	"image"
	"image/color"
	"sort"

	"klabautermann/internal/vision"
)

// Detection holds a class match with spike-grade thresholds until live frames calibrate them.
type Detection struct {
	Class Class
	Score float64
	At    image.Rectangle
}

// DetectClasses scans icons with spike-grade thresholds until live frames calibrate them.
func DetectClasses(frame image.Image, icons map[Class]image.Image, roi image.Rectangle, threshold float64) []Detection {
	if frame == nil {
		return nil
	}
	if roi.Empty() {
		roi = frame.Bounds()
	}
	var out []Detection
	for class, icon := range icons {
		if icon == nil || icon.Bounds().Empty() {
			continue
		}
		m, err := vision.Match(frame, icon, roi)
		if err != nil {
			continue
		}
		if m.Score >= threshold {
			out = append(out, Detection{Class: class, Score: m.Score, At: m.Rect})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Score > out[j].Score })
	return out
}

// GaugeFill reads fill fraction with spike-grade thresholds until live frames calibrate them.
func GaugeFill(frame image.Image, rect image.Rectangle) float64 {
	if frame == nil || rect.Empty() {
		return 0
	}
	clipped := rect.Intersect(frame.Bounds())
	if clipped.Empty() {
		return 0
	}
	var bright, total int
	for y := clipped.Min.Y; y < clipped.Max.Y; y++ {
		for x := clipped.Min.X; x < clipped.Max.X; x++ {
			total++
			if color.GrayModel.Convert(frame.At(x, y)).(color.Gray).Y > 100 {
				bright++
			}
		}
	}
	if total == 0 {
		return 0
	}
	v := float64(bright) / float64(total)
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

type SensedCard struct {
	Kind   CardKind
	Slot   int
	Center image.Point
	Score  float64
}

// DetectCards matches faces resources/templates/card-buster.png, card-arts.png, card-quick.png.
// Slot rects come from live calibration and thresholds are spike-grade.
func DetectCards(frame image.Image, faces map[CardKind]image.Image, slots []image.Rectangle, threshold float64) []SensedCard {
	if frame == nil {
		return nil
	}
	if len(slots) == 0 {
		return nil
	}
	if len(faces) == 0 {
		return nil
	}
	var out []SensedCard
	for i, slot := range slots {
		if slot.Empty() {
			continue
		}
		best := SensedCard{Slot: i}
		matched := false
		for kind, face := range faces {
			if face == nil || face.Bounds().Empty() {
				continue
			}
			m, err := vision.Match(frame, face, slot)
			if err != nil {
				continue
			}
			if m.Score >= threshold && (!matched || m.Score > best.Score) {
				best.Kind = kind
				best.Center = m.Center
				best.Score = m.Score
				matched = true
			}
		}
		if matched {
			out = append(out, best)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Slot < out[j].Slot })
	return out
}
