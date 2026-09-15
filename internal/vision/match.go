package vision

import (
	"errors"
	"image"
	"image/color"
	"math"
)

type MatchResult struct {
	Rect   image.Rectangle
	Center image.Point
	Score  float64
}

func Match(frame, tmpl image.Image, roi image.Rectangle) (MatchResult, error) {
	if frame == nil || frame.Bounds().Empty() {
		return MatchResult{}, errors.New("vision: empty frame bounds")
	}
	if tmpl == nil || tmpl.Bounds().Empty() {
		return MatchResult{}, errors.New("vision: empty template bounds")
	}
	clipped := roi.Intersect(frame.Bounds())
	if clipped.Empty() {
		return MatchResult{}, errors.New("vision: ROI clipped to nothing")
	}
	tb := tmpl.Bounds()
	tw := tb.Dx()
	th := tb.Dy()
	if tw > clipped.Dx() || th > clipped.Dy() {
		return MatchResult{}, errors.New("vision: template larger than clipped ROI")
	}
	fb := frame.Bounds()
	fw := fb.Dx()
	fh := fb.Dy()
	gray := make([]float64, fw*fh)
	for y := fb.Min.Y; y < fb.Max.Y; y++ {
		for x := fb.Min.X; x < fb.Max.X; x++ {
			gray[(y-fb.Min.Y)*fw+(x-fb.Min.X)] = float64(color.GrayModel.Convert(frame.At(x, y)).(color.Gray).Y)
		}
	}
	tvals := make([]float64, tw*th)
	var tsum float64
	for y := tb.Min.Y; y < tb.Max.Y; y++ {
		for x := tb.Min.X; x < tb.Max.X; x++ {
			v := float64(color.GrayModel.Convert(tmpl.At(x, y)).(color.Gray).Y)
			tvals[(y-tb.Min.Y)*tw+(x-tb.Min.X)] = v
			tsum += v
		}
	}
	n := float64(tw * th)
	tmean := tsum / n
	var tvar float64
	for _, v := range tvals {
		d := v - tmean
		tvar += d * d
	}
	if tvar == 0 {
		return MatchResult{}, errors.New("vision: zero-variance template")
	}
	tstd := math.Sqrt(tvar)
	stride := fw + 1
	sum := make([]float64, stride*(fh+1))
	sq := make([]float64, stride*(fh+1))
	for y := 0; y < fh; y++ {
		var row float64
		var rowSq float64
		for x := 0; x < fw; x++ {
			v := gray[y*fw+x]
			row += v
			rowSq += v * v
			sum[(y+1)*stride+x+1] = sum[y*stride+x+1] + row
			sq[(y+1)*stride+x+1] = sq[y*stride+x+1] + rowSq
		}
	}
	best := image.Rectangle{}
	bestScore := -2.0
	for y := clipped.Min.Y; y+th <= clipped.Max.Y; y++ {
		for x := clipped.Min.X; x+tw <= clipped.Max.X; x++ {
			rx0 := x - fb.Min.X
			ry0 := y - fb.Min.Y
			rx1 := rx0 + tw
			ry1 := ry0 + th
			s := sum[ry1*stride+rx1] - sum[ry0*stride+rx1] - sum[ry1*stride+rx0] + sum[ry0*stride+rx0]
			q := sq[ry1*stride+rx1] - sq[ry0*stride+rx1] - sq[ry1*stride+rx0] + sq[ry0*stride+rx0]
			wmean := s / n
			wvar := q - s*s/n
			var score float64
			if wvar <= 0 {
				score = 0
			} else {
				var cov float64
				for ty := 0; ty < th; ty++ {
					base := (ry0+ty)*fw + rx0
					toff := ty * tw
					for tx := 0; tx < tw; tx++ {
						cov += (gray[base+tx] - wmean) * (tvals[toff+tx] - tmean)
					}
				}
				score = cov / (math.Sqrt(wvar) * tstd)
				if score > 1 {
					score = 1
				}
				if score < -1 {
					score = -1
				}
			}
			if score > bestScore {
				best = image.Rect(x, y, x+tw, y+th)
				bestScore = score
			}
		}
	}
	return MatchResult{
		Rect:   best,
		Center: image.Point{X: (best.Min.X + best.Max.X) / 2, Y: (best.Min.Y + best.Max.Y) / 2},
		Score:  bestScore,
	}, nil
}
