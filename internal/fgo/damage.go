package fgo

import "math"

const epsilon = 1e-9

func CardValue(k CardKind) float64 {
	switch k {
	case Buster:
		return 1.5
	case Arts:
		return 1.0
	case Quick:
		return 0.8
	default:
		return 1.0
	}
}

func FirstCardBonus(k CardKind) float64 {
	if k == Buster {
		return 0.5
	}
	return 0
}

func PositionFactor(cardPos int, isNP bool) float64 {
	if isNP {
		return 1.0
	}
	switch {
	case cardPos <= 1:
		return 1.0
	case cardPos == 2:
		return 1.2
	default:
		return 1.4
	}
}

type DamageInput struct {
	Atk        float64
	NPMult     float64
	Card       CardKind
	Position   int
	IsNP       bool
	ClassBonus float64
	Triangle   float64
	Attribute  float64
	AtkMod     float64
	Crit       bool
	Random     float64
}

func Estimate(in DamageInput) int {
	np := in.NPMult
	if np == 0 {
		np = 1.0
	}
	attr := in.Attribute
	if attr == 0 {
		attr = 1.0
	}
	rand := in.Random
	if rand == 0 {
		rand = 1.0
	}
	critMult := 1.0
	if in.Crit {
		critMult = 2.0
	}
	v := in.Atk * np * (FirstCardBonus(in.Card) + CardValue(in.Card)*PositionFactor(in.Position, in.IsNP)) * in.ClassBonus * in.Triangle * attr * rand * 0.23 * math.Max(1+in.AtkMod, 0) * critMult
	out := int(math.Trunc(v * (1 + epsilon)))
	if out < 0 {
		return 0
	}
	return out
}

// 1.3 is bot policy margin covering random spread plus unknown buffs.
func SafeKill(est, hp int) bool {
	return 10*est >= 13*hp
}
