package fgo_test

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"klabautermann/internal/fgo"
)

func TestCombatClearImmediately(t *testing.T) {
	c := &fgo.Combat{}
	attacks, taps := 0, 0
	turns, err := c.Run(
		context.Background(),
		func(context.Context) (bool, []fgo.Card, error) { return true, nil, nil },
		func([]fgo.Card) []fgo.Card { t.Error("pick called after clear"); return nil },
		func(context.Context) error { attacks++; return nil },
		func(context.Context, int, int) error { taps++; return nil },
	)
	if err != nil {
		t.Fatalf("Run = _, %v, want nil", err)
	}
	if turns != 0 {
		t.Errorf("Run turns = %d, want 0", turns)
	}
	if attacks != 0 || taps != 0 {
		t.Errorf("Run attacks = %d taps = %d, want 0 0", attacks, taps)
	}
}

func TestCombatTwoTurnClear(t *testing.T) {
	c := &fgo.Combat{}
	dealA := []fgo.Card{
		{Kind: fgo.Buster, X: 1, Y: 1},
		{Kind: fgo.Arts, X: 2, Y: 2},
		{Kind: fgo.Quick, X: 3, Y: 3},
		{Kind: fgo.Buster, X: 4, Y: 4},
		{Kind: fgo.Arts, X: 5, Y: 5},
	}
	dealB := []fgo.Card{
		{Kind: fgo.Quick, X: 6, Y: 6},
		{Kind: fgo.Buster, X: 7, Y: 7},
		{Kind: fgo.Arts, X: 8, Y: 8},
		{Kind: fgo.Quick, X: 9, Y: 9},
		{Kind: fgo.Buster, X: 10, Y: 10},
	}
	calls := 0
	sense := func(context.Context) (bool, []fgo.Card, error) {
		calls++
		switch calls {
		case 2:
			return false, dealA, nil
		case 4:
			return false, dealB, nil
		case 5:
			return true, nil, nil
		default:
			return false, nil, nil
		}
	}
	pick := func(cards []fgo.Card) []fgo.Card { return cards[:3] }
	attacks := 0
	var tapped [][2]int
	turns, err := c.Run(
		context.Background(),
		sense,
		pick,
		func(context.Context) error { attacks++; return nil },
		func(_ context.Context, x, y int) error { tapped = append(tapped, [2]int{x, y}); return nil },
	)
	if err != nil {
		t.Fatalf("Run = _, %v, want nil", err)
	}
	if turns != 2 {
		t.Errorf("Run turns = %d, want 2", turns)
	}
	if attacks != 2 {
		t.Errorf("Run attacks = %d, want 2", attacks)
	}
	want := [][2]int{{1, 1}, {2, 2}, {3, 3}, {6, 6}, {7, 7}, {8, 8}}
	if !reflect.DeepEqual(tapped, want) {
		t.Errorf("Run taps = %v, want %v", tapped, want)
	}
}

func TestCombatMaxTurns(t *testing.T) {
	c := &fgo.Combat{MaxTurns: 3}
	deal := []fgo.Card{
		{Kind: fgo.Buster, X: 1, Y: 1},
		{Kind: fgo.Arts, X: 2, Y: 2},
		{Kind: fgo.Quick, X: 3, Y: 3},
		{Kind: fgo.Buster, X: 4, Y: 4},
		{Kind: fgo.Arts, X: 5, Y: 5},
	}
	turns, err := c.Run(
		context.Background(),
		func(context.Context) (bool, []fgo.Card, error) { return false, deal, nil },
		nil,
		func(context.Context) error { return nil },
		func(context.Context, int, int) error { return nil },
	)
	if err == nil {
		t.Fatal("Run = _, nil, want max turns error")
	}
	if turns != 3 {
		t.Errorf("Run turns = %d, want 3", turns)
	}
	if !strings.Contains(err.Error(), "3") {
		t.Errorf("Run err = %v, want it to name the limit", err)
	}
}

func TestCombatSenseError(t *testing.T) {
	c := &fgo.Combat{}
	boom := errors.New("no signal")
	turns, err := c.Run(
		context.Background(),
		func(context.Context) (bool, []fgo.Card, error) { return false, nil, boom },
		nil,
		func(context.Context) error { return nil },
		func(context.Context, int, int) error { return nil },
	)
	if !errors.Is(err, boom) {
		t.Fatalf("Run = _, %v, want %v", err, boom)
	}
	if turns != 0 {
		t.Errorf("Run turns = %d, want 0", turns)
	}
}

func TestCombatNilPickUsesBusterFirst(t *testing.T) {
	c := &fgo.Combat{}
	deal := []fgo.Card{
		{Kind: fgo.Quick, X: 1, Y: 0},
		{Kind: fgo.Arts, X: 2, Y: 0},
		{Kind: fgo.Buster, X: 3, Y: 0},
		{Kind: fgo.Arts, X: 4, Y: 0},
		{Kind: fgo.Buster, X: 5, Y: 0},
	}
	calls := 0
	sense := func(context.Context) (bool, []fgo.Card, error) {
		calls++
		if calls == 2 {
			return false, deal, nil
		}
		if calls > 2 {
			return true, nil, nil
		}
		return false, nil, nil
	}
	var tapped [][2]int
	turns, err := c.Run(
		context.Background(),
		sense,
		nil,
		func(context.Context) error { return nil },
		func(_ context.Context, x, y int) error { tapped = append(tapped, [2]int{x, y}); return nil },
	)
	if err != nil {
		t.Fatalf("Run = _, %v, want nil", err)
	}
	if turns != 1 {
		t.Errorf("Run turns = %d, want 1", turns)
	}
	want := [][2]int{{3, 0}, {5, 0}, {2, 0}}
	if !reflect.DeepEqual(tapped, want) {
		t.Errorf("Run taps = %v, want %v", tapped, want)
	}
}

func TestPreferBuster(t *testing.T) {
	hand := []fgo.Card{
		{Kind: fgo.Quick, X: 1, Y: 0},
		{Kind: fgo.Buster, X: 2, Y: 0},
		{Kind: fgo.Arts, X: 3, Y: 0},
		{Kind: fgo.Quick, X: 4, Y: 0},
		{Kind: fgo.Buster, X: 5, Y: 0},
	}
	want := []fgo.Card{
		{Kind: fgo.Buster, X: 2, Y: 0},
		{Kind: fgo.Buster, X: 5, Y: 0},
		{Kind: fgo.Arts, X: 3, Y: 0},
	}
	if got := fgo.PreferBuster(hand); !reflect.DeepEqual(got, want) {
		t.Errorf("PreferBuster = %v, want %v", got, want)
	}
}
