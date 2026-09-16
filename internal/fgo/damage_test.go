package fgo_test

import (
	"testing"

	"klabautermann/internal/fgo"
)

func TestEstimate(t *testing.T) {
	base := fgo.DamageInput{Atk: 10000, Card: fgo.Buster, Position: 1, ClassBonus: 1, Triangle: 1, Attribute: 1}
	cases := []struct {
		name string
		in   fgo.DamageInput
		want int
	}{
		{"10000*(0.5+1.5)*0.23=4600", base, 4600},
		{"10000*(0.5+1.5)*0.23*2.0=9200", withTriangle(base, 2.0), 9200},
		{"10000*3.0*(0.5+1.5)*0.23=13800", withNPMult(base, 3.0), 13800},
		{"10000*(0+0.8*1.4)*0.23=10000*1.12*0.23=2576", withCardPos(base, fgo.Quick, 3), 2576},
		{"10000*(0.5+1.5)*0.23*2(crit)=9200", withCrit(base), 9200},
		{"4600*0.9=4140", withRandom(base, 0.9), 4140},
		{"4600*1.1=5060", withRandom(base, 1.1), 5060},
		{"4600*1.5=6900", withAtkMod(base, 0.5), 6900},
		{"4600*max(1-2,0)=0", withAtkMod(base, -2.0), 0},
		{"10000*(0+1.0*1.2)*0.23=2760", withCardPos(base, fgo.Arts, 2), 2760},
		{"pos9 clamps to 3rd: 10000*(0+0.8*1.4)*0.23=2576", withCardPos(base, fgo.Quick, 9), 2576},
		{"pos0 clamps to 1st: 10000*(0.5+1.5)*0.23=4600", withPos(base, 0), 4600},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := fgo.Estimate(tc.in); got != tc.want {
				t.Fatalf("Estimate = %d, want %d", got, tc.want)
			}
		})
	}
}

func withTriangle(in fgo.DamageInput, v float64) fgo.DamageInput {
	in.Triangle = v
	return in
}

func withNPMult(in fgo.DamageInput, v float64) fgo.DamageInput {
	in.NPMult = v
	return in
}

func withCardPos(in fgo.DamageInput, k fgo.CardKind, pos int) fgo.DamageInput {
	in.Card = k
	in.Position = pos
	return in
}

func withPos(in fgo.DamageInput, pos int) fgo.DamageInput {
	in.Position = pos
	return in
}

func withCrit(in fgo.DamageInput) fgo.DamageInput {
	in.Crit = true
	return in
}

func withRandom(in fgo.DamageInput, v float64) fgo.DamageInput {
	in.Random = v
	return in
}

func withAtkMod(in fgo.DamageInput, v float64) fgo.DamageInput {
	in.AtkMod = v
	return in
}

func TestSafeKill(t *testing.T) {
	if !fgo.SafeKill(4600, 3000) {
		t.Fatal("SafeKill(4600,3000) = false, want true: 4600 >= 1.3*3000=3900")
	}
	if fgo.SafeKill(4600, 4000) {
		t.Fatal("SafeKill(4600,4000) = true, want false: 4600 < 1.3*4000=5200")
	}
}
