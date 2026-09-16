package fgo_test

import (
	"testing"

	"klabautermann/internal/fgo"
)

func TestAdvantage(t *testing.T) {
	cases := []struct {
		atk  fgo.Class
		def  fgo.Class
		want float64
	}{
		{fgo.Saber, fgo.Archer, 0.5},
		{fgo.Saber, fgo.Lancer, 2.0},
		{fgo.Saber, fgo.Berserker, 2.0},
		{fgo.Saber, fgo.Ruler, 0.5},
		{fgo.Archer, fgo.Saber, 2.0},
		{fgo.Archer, fgo.Lancer, 0.5},
		{fgo.Archer, fgo.Berserker, 2.0},
		{fgo.Archer, fgo.Ruler, 0.5},
		{fgo.Lancer, fgo.Saber, 0.5},
		{fgo.Lancer, fgo.Archer, 2.0},
		{fgo.Lancer, fgo.Berserker, 2.0},
		{fgo.Lancer, fgo.Ruler, 0.5},
		{fgo.Rider, fgo.Caster, 2.0},
		{fgo.Rider, fgo.Assassin, 0.5},
		{fgo.Rider, fgo.Berserker, 2.0},
		{fgo.Rider, fgo.Ruler, 0.5},
		{fgo.Caster, fgo.Caster, 0.5},
		{fgo.Caster, fgo.Assassin, 2.0},
		{fgo.Caster, fgo.Berserker, 2.0},
		{fgo.Caster, fgo.Ruler, 0.5},
		{fgo.Assassin, fgo.Rider, 2.0},
		{fgo.Assassin, fgo.Caster, 0.5},
		{fgo.Assassin, fgo.Berserker, 2.0},
		{fgo.Assassin, fgo.Ruler, 0.5},
		{fgo.Berserker, fgo.Saber, 1.5},
		{fgo.Berserker, fgo.Archer, 1.5},
		{fgo.Berserker, fgo.Lancer, 1.5},
		{fgo.Berserker, fgo.Rider, 1.5},
		{fgo.Berserker, fgo.Caster, 1.5},
		{fgo.Berserker, fgo.Assassin, 1.5},
		{fgo.Berserker, fgo.Berserker, 1.5},
		{fgo.Berserker, fgo.Ruler, 1.5},
		{fgo.Berserker, fgo.Avenger, 1.5},
		{fgo.Berserker, fgo.MoonCancer, 1.5},
		{fgo.Berserker, fgo.AlterEgo, 1.5},
		{fgo.Berserker, fgo.Foreigner, 0.5},
		{fgo.Berserker, fgo.Pretender, 1.5},
		{fgo.Berserker, fgo.Beast, 1.5},
		{fgo.Ruler, fgo.Berserker, 2.0},
		{fgo.Ruler, fgo.Avenger, 0.5},
		{fgo.Ruler, fgo.MoonCancer, 2.0},
		{fgo.Avenger, fgo.Ruler, 2.0},
		{fgo.Avenger, fgo.MoonCancer, 0.5},
		{fgo.Avenger, fgo.Berserker, 2.0},
		{fgo.MoonCancer, fgo.Ruler, 0.5},
		{fgo.MoonCancer, fgo.Avenger, 2.0},
		{fgo.MoonCancer, fgo.Berserker, 2.0},
		{fgo.AlterEgo, fgo.Saber, 0.5},
		{fgo.AlterEgo, fgo.Archer, 0.5},
		{fgo.AlterEgo, fgo.Lancer, 0.5},
		{fgo.AlterEgo, fgo.Rider, 1.5},
		{fgo.AlterEgo, fgo.Caster, 1.5},
		{fgo.AlterEgo, fgo.Assassin, 1.5},
		{fgo.AlterEgo, fgo.Berserker, 2.0},
		{fgo.AlterEgo, fgo.Foreigner, 2.0},
		{fgo.AlterEgo, fgo.Pretender, 0.5},
		{fgo.Foreigner, fgo.AlterEgo, 0.5},
		{fgo.Foreigner, fgo.Foreigner, 2.0},
		{fgo.Foreigner, fgo.Pretender, 2.0},
		{fgo.Foreigner, fgo.Berserker, 2.0},
		{fgo.Pretender, fgo.Saber, 1.5},
		{fgo.Pretender, fgo.Archer, 1.5},
		{fgo.Pretender, fgo.Lancer, 1.5},
		{fgo.Pretender, fgo.Rider, 0.5},
		{fgo.Pretender, fgo.Caster, 0.5},
		{fgo.Pretender, fgo.Assassin, 0.5},
		{fgo.Pretender, fgo.Berserker, 2.0},
		{fgo.Pretender, fgo.AlterEgo, 2.0},
		{fgo.Pretender, fgo.Foreigner, 0.5},
		{fgo.Beast, fgo.Saber, 2.0},
		{fgo.Beast, fgo.Archer, 2.0},
		{fgo.Beast, fgo.Lancer, 2.0},
		{fgo.Beast, fgo.Berserker, 2.0},
		{fgo.Beast, fgo.Avenger, 0.5},
		{fgo.Saber, fgo.Rider, 1.0},
		{fgo.Shielder, fgo.Foreigner, 1.0},
	}
	for _, tc := range cases {
		if got := fgo.Advantage(tc.atk, tc.def); got != tc.want {
			t.Errorf("Advantage(%q, %q) = %v, want %v", tc.atk, tc.def, got, tc.want)
		}
	}
}

func TestAdvantageComplete(t *testing.T) {
	classes := []fgo.Class{
		fgo.Shielder, fgo.Saber, fgo.Archer, fgo.Lancer, fgo.Rider,
		fgo.Caster, fgo.Assassin, fgo.Berserker, fgo.Ruler, fgo.Avenger,
		fgo.MoonCancer, fgo.AlterEgo, fgo.Foreigner, fgo.Pretender, fgo.Beast,
	}
	allowed := map[float64]bool{0.5: true, 1.0: true, 1.2: true, 1.5: true, 2.0: true}
	for _, atk := range classes {
		if !atk.Valid() {
			t.Errorf("Valid(%q) = false, want true", atk)
		}
		for _, def := range classes {
			if got := fgo.Advantage(atk, def); !allowed[got] {
				t.Errorf("Advantage(%q, %q) = %v, want one of 0.5, 1.0, 1.2, 1.5, 2.0", atk, def, got)
			}
		}
	}
}

func TestClassBase(t *testing.T) {
	cases := []struct {
		class fgo.Class
		want  float64
	}{
		{fgo.Saber, 1.0},
		{fgo.Archer, 0.95},
		{fgo.Lancer, 1.05},
		{fgo.Rider, 1.0},
		{fgo.Caster, 0.9},
		{fgo.Assassin, 0.9},
		{fgo.Berserker, 1.1},
		{fgo.Shielder, 1.0},
		{fgo.Ruler, 1.1},
		{fgo.Avenger, 1.1},
		{fgo.MoonCancer, 1.0},
		{fgo.AlterEgo, 1.0},
		{fgo.Foreigner, 1.0},
		{fgo.Pretender, 1.0},
		{fgo.Beast, 1.0},
	}
	for _, tc := range cases {
		if got := fgo.ClassBase(tc.class); got != tc.want {
			t.Errorf("ClassBase(%q) = %v, want %v", tc.class, got, tc.want)
		}
	}
}

func TestUnknownClass(t *testing.T) {
	if got := fgo.Advantage(fgo.Class("XX"), fgo.Saber); got != 1.0 {
		t.Errorf("Advantage(%q, %q) = %v, want 1.0", "XX", fgo.Saber, got)
	}
	if got := fgo.Advantage(fgo.Saber, fgo.Class("XX")); got != 1.0 {
		t.Errorf("Advantage(%q, %q) = %v, want 1.0", fgo.Saber, "XX", got)
	}
	if got := fgo.ClassBase(fgo.Class("XX")); got != 1.0 {
		t.Errorf("ClassBase(%q) = %v, want 1.0", "XX", got)
	}
}
