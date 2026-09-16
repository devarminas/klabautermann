package fgo_test

import (
	"reflect"
	"testing"

	"klabautermann/internal/fgo"
)

func TestClassOfNode(t *testing.T) {
	cases := map[string]fgo.Class{
		"ClassMoonCancer": fgo.MoonCancer,
		"ClassCaster":     fgo.Caster,
		"ClassPretender":  fgo.Pretender,
		"ClassRider":      fgo.Rider,
		"ClassShielder":   fgo.Shielder,
		"EnemyRider":      fgo.Rider,
		"EnemySaber":      fgo.Saber,
	}
	for name, want := range cases {
		got, ok := fgo.ClassOfNode(name)
		if !ok || got != want {
			t.Errorf("ClassOfNode(%q) = %q, %v; want %q, true", name, got, ok, want)
		}
	}
	if got, ok := fgo.ClassOfNode("Nope"); ok || got != "" {
		t.Errorf("ClassOfNode(%q) = %q, %v; want %q, false", "Nope", got, ok, "")
	}
}

func TestBestAllyLiveWave(t *testing.T) {
	enemies := []fgo.Class{fgo.Rider, fgo.Rider, fgo.Saber}
	allies := []fgo.Class{fgo.Pretender, fgo.Caster, fgo.MoonCancer}
	got, score := fgo.BestAlly(enemies, allies)
	if got != fgo.Caster || score != 3.0 {
		t.Fatalf("BestAlly = %q, %v; want %q, 3.0", got, score, fgo.Caster)
	}
	if s := fgo.Advantage(fgo.Caster, fgo.Rider); s != 1.0 {
		t.Errorf("Advantage(Caster,Rider) = %v; want 1.0", s)
	}
	if s := fgo.Advantage(fgo.Caster, fgo.Saber); s != 1.0 {
		t.Errorf("Advantage(Caster,Saber) = %v; want 1.0", s)
	}
	tied, tiedScore := fgo.BestAlly(enemies, []fgo.Class{fgo.MoonCancer, fgo.Caster})
	if tied != fgo.MoonCancer || tiedScore != 3.0 {
		t.Errorf("tie BestAlly = %q, %v; want MoonCancer, 3.0", tied, tiedScore)
	}
}

func TestWeakTo(t *testing.T) {
	all := []fgo.Class{fgo.Saber, fgo.Archer, fgo.Lancer, fgo.Rider, fgo.Caster, fgo.Assassin, fgo.Berserker, fgo.Shielder, fgo.Ruler, fgo.Avenger, fgo.MoonCancer, fgo.AlterEgo, fgo.Foreigner, fgo.Pretender, fgo.Beast}
	if got := fgo.WeakTo(fgo.Rider, all); !reflect.DeepEqual(got, []fgo.Class{fgo.Assassin}) {
		t.Errorf("WeakTo(Rider) = %v; want [Assassin]", got)
	}
	if got := fgo.WeakTo(fgo.Assassin, all); !reflect.DeepEqual(got, []fgo.Class{fgo.Caster}) {
		t.Errorf("WeakTo(Assassin) = %v; want [Caster]", got)
	}
	if s := fgo.Advantage(fgo.Assassin, fgo.Rider); s != 2.0 {
		t.Errorf("Advantage(Assassin,Rider) = %v; want 2.0", s)
	}
	if s := fgo.Advantage(fgo.Caster, fgo.Assassin); s != 2.0 {
		t.Errorf("Advantage(Caster,Assassin) = %v; want 2.0", s)
	}
}

func TestBestAllyEmpty(t *testing.T) {
	got, score := fgo.BestAlly([]fgo.Class{fgo.Rider}, nil)
	if got != "" || score != 0 {
		t.Errorf("BestAlly empty = %q, %v; want %q, 0", got, score, "")
	}
	if got := fgo.WeakTo(fgo.Rider, nil); len(got) != 0 {
		t.Errorf("WeakTo empty = %v; want []", got)
	}
}
